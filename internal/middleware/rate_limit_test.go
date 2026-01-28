package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddlewareMemory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		rps        float64
		burst      int
		reqCount   int
		expectCode int
	}{
		{
			name:       "Allow requests under limit",
			rps:        10,
			burst:      10,
			reqCount:   5,
			expectCode: http.StatusOK,
		},
		{
			name:       "Block requests over limit",
			rps:        1,
			burst:      1,
			reqCount:   3,
			expectCode: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(RateLimitMiddlewareMemory(tt.rps, tt.burst))
			r.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			for i := 0; i < tt.reqCount; i++ {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/", nil)
				r.ServeHTTP(w, req)

				if i < int(tt.rps) && tt.rps > 1 {
					assert.Equal(t, http.StatusOK, w.Code)
				} else if i >= int(tt.burst) && tt.burst == 1 {
					// For the blocking case
					if i == 0 {
						assert.Equal(t, http.StatusOK, w.Code)
					} else {
						assert.Equal(t, http.StatusTooManyRequests, w.Code)
					}
				}
			}
		})
	}
}

func TestRateLimitMiddlewareRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := logrus.New()

	t.Run("Allow requests", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		// When using Lua script, redismock handles Eval/EvalSha
		// The key format is "rate_limit:ip:"

		// First request: Script returns 1
		mock.ExpectEvalSha(rateLimitScript.Hash(), []string{"rate_limit:ip:"}, 60).SetVal(int64(1))

		// Second request: Script returns 2
		mock.ExpectEvalSha(rateLimitScript.Hash(), []string{"rate_limit:ip:"}, 60).SetVal(int64(2))

		r := gin.New()
		r.Use(RateLimitMiddlewareRedis(db, logger, LimiterTypeIP, 10, 60*time.Second))
		r.GET("/", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		// Request 1
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Request 2
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Block requests", func(t *testing.T) {
		db, mock := redismock.NewClientMock()

		limit := 60

		// Mock hitting the limit
		// Script returns limit + 1
		mock.ExpectEvalSha(rateLimitScript.Hash(), []string{"rate_limit:ip:"}, 60).SetVal(int64(limit + 1))

		r := gin.New()
		r.Use(RateLimitMiddlewareRedis(db, logger, LimiterTypeIP, limit, 60*time.Second))
		r.GET("/", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
