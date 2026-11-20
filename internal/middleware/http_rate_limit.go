package middleware

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type LimiterType string

const (
	LimiterTypeIP   LimiterType = "ip"
	LimiterTypeUser LimiterType = "user"
)

// rateLimitScript is a Lua script to atomically increment and set expiry if needed.
// KEYS[1]: rate limit key
// ARGV[1]: window in seconds
var rateLimitScript = redis.NewScript(`
	local current = redis.call("INCR", KEYS[1])
	if current == 1 then
		redis.call("EXPIRE", KEYS[1], ARGV[1])
	end
	return current
`)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	ips map[string]*clientLimiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*clientLimiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	entry, exists := i.ips[ip]
	if !exists {
		entry = &clientLimiter{
			limiter: rate.NewLimiter(i.r, i.b),
		}
		i.ips[ip] = entry
	}
	entry.lastSeen = time.Now()

	return entry.limiter
}

type clientIPCtxKey struct{}

// HTTPTrustedProxyMiddleware resolves the real client IP from X-Forwarded-For or X-Real-IP
// only when the immediate peer address matches one of the configured trusted proxy IPs/CIDRs.
func HTTPTrustedProxyMiddleware(trustedProxies []string, log *logrus.Logger) delivery.Middleware {
	var prefixes []netip.Prefix
	for _, entry := range trustedProxies {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			p, err := netip.ParsePrefix(entry)
			if err != nil {
				if log != nil {
					log.WithError(err).Warnf("Ignoring invalid trusted proxy CIDR: %s", entry)
				}
				continue
			}
			prefixes = append(prefixes, p.Masked())
		} else {
			addr, err := netip.ParseAddr(entry)
			if err != nil {
				if log != nil {
					log.WithError(err).Warnf("Ignoring invalid trusted proxy IP: %s", entry)
				}
				continue
			}
			addr = addr.Unmap()
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
		}
	}

	if len(prefixes) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	isTrusted := func(ipStr string) bool {
		addr, err := netip.ParseAddr(strings.TrimSpace(ipStr))
		if err != nil {
			return false
		}
		addr = addr.Unmap()
		for _, p := range prefixes {
			if p.Contains(addr) {
				return true
			}
		}
		return false
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			remoteIP := remoteHostIP(r.RemoteAddr)
			if remoteIP != "" && isTrusted(remoteIP) {
				resolved := ""
				if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
					parts := strings.Split(xff, ",")
					for i := len(parts) - 1; i >= 0; i-- {
						candidate := strings.TrimSpace(parts[i])
						if candidate == "" {
							continue
						}
						if !isTrusted(candidate) {
							resolved = candidate
							break
						}
					}
					if resolved == "" && len(parts) > 0 {
						resolved = strings.TrimSpace(parts[0])
					}
				} else if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
					resolved = xri
				}

				if resolved != "" {
					ctx := context.WithValue(r.Context(), clientIPCtxKey{}, resolved)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func remoteHostIP(remoteAddr string) string {
	addr := strings.TrimSpace(remoteAddr)
	if addr == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

func extractClientIP(r *http.Request) string {
	if resolved, ok := r.Context().Value(clientIPCtxKey{}).(string); ok && resolved != "" {
		return resolved
	}
	return remoteHostIP(r.RemoteAddr)
}

func isLocalhostIP(ip string) bool {
	return ip == "127.0.0.1" || ip == "::1"
}

// HTTPRateLimitMiddlewareRedis implements a Redis-backed fixed-window rate limiter for net/http.
func HTTPRateLimitMiddlewareRedis(rdb *redis.Client, log *logrus.Logger, limiterType LimiterType, limit int, windowSeconds int) delivery.Middleware {
	if limit < 1 {
		limit = 1
	}
	if windowSeconds < 1 {
		windowSeconds = 60
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rdb == nil {
				next.ServeHTTP(w, r)
				return
			}

			clientIP := extractClientIP(r)
			if isLocalhostIP(clientIP) {
				next.ServeHTTP(w, r)
				return
			}

			var identifier string
			var key string

			if limiterType == LimiterTypeUser {
				if userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey); ok {
					identifier = userID
					key = fmt.Sprintf("rate_limit:%s:%s", limiterType, identifier)
				} else {
					identifier = clientIP
					key = fmt.Sprintf("rate_limit:%s:%s", "ip_fallback", identifier)
				}
			} else {
				identifier = clientIP
				key = fmt.Sprintf("rate_limit:%s:%s", limiterType, identifier)
			}

			count, err := rateLimitScript.Run(r.Context(), rdb, []string{key}, windowSeconds).Int64()
			if err != nil {
				if log != nil {
					log.Errorf("Rate limit redis error: %v", err)
				}
				next.ServeHTTP(w, r)
				return
			}

			remaining := int64(limit) - count
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

			if count > int64(limit) {
				telemetry.RecordRateLimitRejection(string(limiterType))
				w.Header().Set("Retry-After", strconv.Itoa(windowSeconds))
				if log != nil {
					log.Warnf("Rate limit exceeded for %s: %s (Count: %d, Limit: %d)", limiterType, identifier, count, limit)
				}
				response.WriteError(w, http.StatusTooManyRequests, exception.ErrTooManyRequests, "Too many requests, please try again later.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// HTTPRateLimitMiddlewareMemory implements an in-memory per-IP token bucket rate limiter for net/http.
func HTTPRateLimitMiddlewareMemory(rps float64, burst int) delivery.Middleware {
	limiter := NewIPRateLimiter(rate.Limit(rps), burst)

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.mu.Lock()
			for ip, client := range limiter.ips {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(limiter.ips, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractClientIP(r)
			if isLocalhostIP(ip) {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(burst))

			ipLimiter := limiter.GetLimiter(ip)
			if !ipLimiter.Allow() {
				telemetry.RecordRateLimitRejection("memory")
				retryAfter := 1
				if rps > 0 && rps < 1 {
					retryAfter = int(math.Ceil(1.0 / rps))
				}
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				response.WriteError(w, http.StatusTooManyRequests, exception.ErrTooManyRequests, "rate limit exceeded")
				return
			}

			remaining := int(ipLimiter.Tokens())
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			next.ServeHTTP(w, r)
		})
	}
}
