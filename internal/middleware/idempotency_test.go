package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIdempotencyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)

	t.Run("First request - Success and Cache", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		router := gin.New()
		router.Use(Idempotency(rdb, log, 10*time.Second))

		router.POST("/submit", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"counter": 1})
		})

		bodyStr := `{"data": 123}`
		hasher := sha256.New()
		hasher.Write([]byte(bodyStr))
		payloadHash := hex.EncodeToString(hasher.Sum(nil))

		mock.ExpectSetNX("idempotency_lock:key-1", payloadHash, 10*time.Second).SetVal(true)
		mock.ExpectGet("idempotency:key-1").RedisNil()
		mock.ExpectDel("idempotency_lock:key-1").SetVal(1)

		respToCache := IdempotencyResponse{
			Status:  http.StatusOK,
			Headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
			Body:    []byte(`{"counter":1}`),
		}
		cachedBytes, _ := json.Marshal(respToCache)
		mock.ExpectSet("idempotency:key-1", string(cachedBytes), 10*time.Second).SetVal("OK")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(bodyStr))
		req.Header.Set("Idempotency-Key", "key-1")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Second request - Replays Cached Response", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		router := gin.New()
		router.Use(Idempotency(rdb, log, 10*time.Second))

		router.POST("/submit", func(c *gin.Context) {
			t.Error("handler should not be called")
		})

		bodyStr := `{"data": 123}`
		hasher := sha256.New()
		hasher.Write([]byte(bodyStr))
		payloadHash := hex.EncodeToString(hasher.Sum(nil))

		mock.ExpectSetNX("idempotency_lock:key-1", payloadHash, 10*time.Second).SetVal(false)
		mock.ExpectGet("idempotency_lock:key-1").SetVal(payloadHash)

		respToCache := IdempotencyResponse{
			Status:  http.StatusOK,
			Headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
			Body:    []byte(`{"counter":1}`),
		}
		cachedBytes, _ := json.Marshal(respToCache)
		mock.ExpectGet("idempotency:key-1").SetVal(string(cachedBytes))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(bodyStr))
		req.Header.Set("Idempotency-Key", "key-1")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "HIT - Idempotency Replay", w.Header().Get("X-Cache-Lookup"))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
