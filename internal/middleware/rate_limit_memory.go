package middleware

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

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

func RateLimitMiddlewareMemory(rps float64, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(rps), burst)

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			limiter.mu.Lock()
			for ip, client := range limiter.ips {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(limiter.ips, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			response.ErrorResponse(c, http.StatusTooManyRequests, errors.New("too many requests"), "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}
