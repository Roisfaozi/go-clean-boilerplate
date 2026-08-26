package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type IdempotencyResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Idempotency(rdb *redis.Client, log *logrus.Logger, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check write methods (POST, PUT, PATCH)
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		// Prevent duplicate processing with client payload hash check
		var bodyBytes []byte
		if c.Request.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		hasher := sha256.New()
		hasher.Write(bodyBytes)
		payloadHash := hex.EncodeToString(hasher.Sum(nil))

		redisKey := "idempotency:" + key
		lockKey := "idempotency_lock:" + key

		// Try to acquire processing lock
		acquired, err := rdb.SetNX(c.Request.Context(), lockKey, payloadHash, 10*time.Second).Result()
		if err != nil {
			log.WithError(err).Error("Idempotency middleware: failed to acquire lock")
			c.Next()
			return
		}

		if !acquired {
			// Check if payload matches existing processing lock
			existingHash, _ := rdb.Get(c.Request.Context(), lockKey).Result()
			if existingHash != payloadHash {
				response.Error(c, http.StatusConflict, errors.New("idempotency key conflict"), "payload mismatch for same idempotency key")
				c.Abort()
				return
			}

			deadline := time.Now().Add(2 * time.Second)
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()

			for time.Now().Before(deadline) {
				cachedData, cacheErr := rdb.Get(c.Request.Context(), redisKey).Result()
				if cacheErr == nil && cachedData != "" {
					var cachedResp IdempotencyResponse
					if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
						for k, v := range cachedResp.Headers {
							c.Header(k, v)
						}
						c.Header("X-Cache-Lookup", "HIT - Idempotency Replay")
						c.Data(cachedResp.Status, "application/json", cachedResp.Body)
						c.Abort()
						return
					}
				}

				select {
				case <-c.Request.Context().Done():
					response.Error(c, http.StatusRequestTimeout, errors.New("idempotency wait cancelled"), "request cancelled while waiting for idempotent replay")
					c.Abort()
					return
				case <-ticker.C:
				}
			}

			response.Error(c, http.StatusConflict, errors.New("idempotency processing timeout"), "original request still processing")
			c.Abort()
			return
		}

		defer rdb.Del(c.Request.Context(), lockKey)

		// Check if we have cached response
		cachedData, err := rdb.Get(c.Request.Context(), redisKey).Result()
		if err == nil && cachedData != "" {
			var cachedResp IdempotencyResponse
			if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
				// Replay response
				for k, v := range cachedResp.Headers {
					c.Header(k, v)
				}
				c.Header("X-Cache-Lookup", "HIT - Idempotency Replay")
				c.Data(cachedResp.Status, "application/json", cachedResp.Body)
				c.Abort()
				return
			}
		}

		// Hook response writer to capture output
		w := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		// Save response to cache if status code is success (2xx)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			headers := make(map[string]string)
			// Capture custom header if necessary, or just Content-Type
			if contentType := c.Writer.Header().Get("Content-Type"); contentType != "" {
				headers["Content-Type"] = contentType
			}

			respToCache := IdempotencyResponse{
				Status:  c.Writer.Status(),
				Headers: headers,
				Body:    w.body.Bytes(),
			}

			cachedBytes, err := json.Marshal(respToCache)
			if err == nil {
				rdb.Set(context.Background(), redisKey, cachedBytes, ttl)
			}
		}
	}
}
