package middleware

import (
	"fmt"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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

// RateLimitMiddlewareRedis implements a simple fixed window rate limiter using Redis.
// It converts the RPS (Requests Per Second) config into a 1-minute fixed window limit.
func RateLimitMiddlewareRedis(redisClient *redis.Client, log *logrus.Logger, rps float64) gin.HandlerFunc {
	limit := int64(rps * 60)
	if limit < 1 {
		limit = 1
	}
	// Window is 60 seconds
	windowSeconds := 60

	return func(c *gin.Context) {
		if redisClient == nil {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Execute Lua script
		count, err := rateLimitScript.Run(c.Request.Context(), redisClient, []string{key}, windowSeconds).Int64()
		if err != nil {
			log.Errorf("Rate limit redis error: %v", err)
			c.Next()
			return
		}

		if count > limit {
			log.Warnf("Rate limit exceeded for IP: %s (Count: %d, Limit: %d)", clientIP, count, limit)
			response.ErrorResponse(c, http.StatusTooManyRequests, exception.ErrTooManyRequests, "Too many requests, please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}
