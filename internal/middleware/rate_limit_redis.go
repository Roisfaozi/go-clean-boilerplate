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

// RateLimitMiddlewareRedis implements a simple fixed window rate limiter using Redis.
// It converts the RPS (Requests Per Second) config into a 1-minute fixed window limit.
func RateLimitMiddlewareRedis(redisClient *redis.Client, log *logrus.Logger, rps float64) gin.HandlerFunc {
	// Convert RPS to requests per minute
	limit := int64(rps * 60)
	if limit < 1 {
		limit = 1
	}
	window := 1 * time.Minute

	return func(c *gin.Context) {
		if redisClient == nil {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		// Increment the counter
		count, err := redisClient.Incr(c.Request.Context(), key).Result()
		if err != nil {
			log.Errorf("Rate limit redis error: %v", err)
			// Fail open: allow request if Redis is down
			c.Next()
			return
		}

		// Set expiration on first request in the window if no TTL exists
		if count == 1 {
			redisClient.Expire(c.Request.Context(), key, window)
		} else {
			// Ensure TTL is set (handle edge case where Incr happens but Expire failed previously)
			ttl, _ := redisClient.TTL(c.Request.Context(), key).Result()
			if ttl == -1 {
				redisClient.Expire(c.Request.Context(), key, window)
			}
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
