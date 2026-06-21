package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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

// RateLimitMiddlewareRedis implements a flexible rate limiter using Redis.
func RateLimitMiddlewareRedis(redisClient *redis.Client, log *logrus.Logger, limitType LimiterType, limit int, window time.Duration) gin.HandlerFunc {
	if limit < 1 {
		limit = 1
	}
	windowSeconds := int(window.Seconds())

	return func(c *gin.Context) {
		if redisClient == nil {
			c.Next()
			return
		}

		// Whitelist localhost for seeding/internal tools
		if c.ClientIP() == "127.0.0.1" || c.ClientIP() == "::1" {
			c.Next()
			return
		}

		var identifier string
		var key string

		if limitType == LimiterTypeUser {
			userID, exists := c.Get("userID")
			if !exists {
				// Fallback to IP if userID not found (e.g. auth failed)
				identifier = c.ClientIP()
				key = fmt.Sprintf("rate_limit:%s:%s", "ip_fallback", identifier)
			} else {
				identifier = userID.(string)
				key = fmt.Sprintf("rate_limit:%s:%s", limitType, identifier)
			}
		} else {
			identifier = c.ClientIP()
			key = fmt.Sprintf("rate_limit:%s:%s", limitType, identifier)
		}

		// Execute Lua script
		count, err := rateLimitScript.Run(c.Request.Context(), redisClient, []string{key}, windowSeconds).Int64()
		if err != nil {
			log.Errorf("Rate limit redis error: %v", err)
			c.Next()
			return
		}

		if count > int64(limit) {
			log.Warnf("Rate limit exceeded for %s: %s (Count: %d, Limit: %d)", limitType, identifier, count, limit)
			response.ErrorResponse(c, http.StatusTooManyRequests, exception.ErrTooManyRequests, "Too many requests, please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}
