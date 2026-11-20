package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	apiKeyModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	apiKeyMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	authModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/model"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/constants"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type httpMockCasbinEnforcer struct {
	mock.Mock
}

func (m *httpMockCasbinEnforcer) Enforce(rvals ...any) (bool, error) {
	args := m.Called(rvals...)
	return args.Bool(0), args.Error(1)
}

func newDiscardLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

func TestHTTPRequestIDMiddleware(t *testing.T) {
	mw := HTTPRequestIDMiddleware()

	t.Run("generates request ID when missing", func(t *testing.T) {
		var ctxReqID string
		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxReqID, _ = r.Context().Value(constants.RequestIDKey).(string)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		headerID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, headerID)
		assert.Equal(t, headerID, ctxReqID)
	})

	t.Run("preserves existing request ID", func(t *testing.T) {
		var ctxReqID string
		h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxReqID, _ = r.Context().Value(constants.RequestIDKey).(string)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "custom-req-123")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, "custom-req-123", w.Header().Get("X-Request-ID"))
		assert.Equal(t, "custom-req-123", ctxReqID)
	})
}

func TestHTTPRequestLogger(t *testing.T) {
	t.Run("logs route, auth_method, response_bytes, and preserves Flusher and Unwrap", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)
		log.SetFormatter(&logrus.JSONFormatter{})

		mux := http.NewServeMux()
		var flusherOK bool
		var unwrappedOK bool

		mux.Handle("GET /api/v1/items/{id}", HTTPRequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, flusherOK = w.(http.Flusher)
			if u, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
				unwrappedOK = u.Unwrap() != nil
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"ok":true}`))
		})))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/items/42", nil)
		ctx := delivery.SetContextValue(req.Context(), delivery.AuthMethodKey, "jwt")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.True(t, flusherOK, "responseRecorder must satisfy http.Flusher")
		assert.True(t, unwrappedOK, "responseRecorder must satisfy Unwrap")

		var entry map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
		assert.Equal(t, "http_request", entry["type"])
		assert.Equal(t, "GET /api/v1/items/{id}", entry["route"])
		assert.Equal(t, "jwt", entry["auth_method"])
		assert.Equal(t, float64(http.StatusCreated), entry["status"])
		assert.Equal(t, float64(len(`{"ok":true}`)), entry["response_bytes"])
		assert.Equal(t, float64(len(`{"ok":true}`)), entry["data_length"])
	})

	t.Run("falls back to unknown route and logs 4xx/5xx levels", func(t *testing.T) {
		for _, tc := range []struct {
			code  int
			level string
		}{
			{http.StatusBadRequest, "warning"},
			{http.StatusInternalServerError, "error"},
		} {
			buf := &bytes.Buffer{}
			log := logrus.New()
			log.SetOutput(buf)
			log.SetFormatter(&logrus.JSONFormatter{})

			h := HTTPRequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.code)
			}))

			req := httptest.NewRequest(http.MethodGet, "/direct", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			var entry map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
			assert.Equal(t, "unknown", entry["route"])
			assert.Equal(t, tc.level, entry["level"])
		}
	})
}

func TestHTTPRecoveryMiddleware(t *testing.T) {
	t.Run("recovers panic before header written and logs with context", func(t *testing.T) {
		buf := &bytes.Buffer{}
		log := logrus.New()
		log.SetOutput(buf)
		log.SetFormatter(&logrus.JSONFormatter{})

		h := delivery.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("boom")
		}), HTTPRequestIDMiddleware(), HTTPRecoveryMiddleware(log))

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		req.Header.Set("X-Request-ID", "panic-req-1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "internal server error")

		var entry map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
		assert.Equal(t, "panic_recovery", entry["type"])
		assert.Equal(t, "panic-req-1", entry["request_id"])
	})

	t.Run("does not overwrite response when headers already written before panic", func(t *testing.T) {
		log := newDiscardLogger()
		h := HTTPRecoveryMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte("partial"))
			panic("late panic")
		}))

		req := httptest.NewRequest(http.MethodGet, "/late-panic", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.Equal(t, "partial", w.Body.String())
	})
}

func TestHTTPSecurityMiddleware(t *testing.T) {
	h := HTTPSecurityMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/sec", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.NotEmpty(t, w.Header().Get("Permissions-Policy"))
}

func TestHTTPCORSMiddleware(t *testing.T) {
	t.Run("handles OPTIONS preflight for allowed origin with 204", func(t *testing.T) {
		nextCalled := false
		h := HTTPCORSMiddleware([]string{"https://app.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		}))

		req := httptest.NewRequest(http.MethodOptions, "/api/v1/projects", nil)
		req.Header.Set("Origin", "https://app.example.com")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.False(t, nextCalled)
		assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("wildcard origin sets star without credentials", func(t *testing.T) {
		h := HTTPCORSMiddleware([]string{"*"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		req.Header.Set("Origin", "https://any.example.com")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("disallowed origin omits Access-Control-Allow-Origin", func(t *testing.T) {
		h := HTTPCORSMiddleware([]string{"https://allowed.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestHTTPCSRFMiddleware(t *testing.T) {
	log := newDiscardLogger()
	h := HTTPCSRFMiddleware(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("safe method GET bypasses CSRF", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Bearer auth bypasses CSRF", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", nil)
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("API key header bypasses CSRF", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", nil)
		req.Header.Set("X-API-Key", "sk_live_test")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("matching cookie and header passes CSRF", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "secret-csrf"})
		req.Header.Set("X-CSRF-Token", "secret-csrf")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing or mismatched token returns 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "secret-csrf"})
		req.Header.Set("X-CSRF-Token", "wrong-csrf")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestHTTPPrometheusMiddleware(t *testing.T) {
	before := GetTotalRequests()
	var flusherOK bool
	var unwrapOK bool

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/prom/{id}", HTTPPrometheusMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, flusherOK = w.(http.Flusher)
		if u, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
			unwrapOK = u.Unwrap() != nil
		}
		w.WriteHeader(http.StatusCreated)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/prom/99", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.True(t, flusherOK)
	assert.True(t, unwrapOK)
	assert.Equal(t, before+1, GetTotalRequests())
}

func TestHTTPRateLimitMiddlewareRedis(t *testing.T) {
	log := newDiscardLogger()

	t.Run("nil redis client passes through", func(t *testing.T) {
		h := HTTPRateLimitMiddlewareRedis(nil, log, LimiterTypeIP, 5, 60)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("localhost is whitelisted", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		h := HTTPRateLimitMiddlewareRedis(rdb, log, LimiterTypeIP, 1, 60)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:54321"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		require.NoError(t, redMock.ExpectationsWereMet())
	})

	t.Run("allows within limit and sets headers", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		redMock.ExpectEvalSha(rateLimitScript.Hash(), []string{"rate_limit:ip:203.0.113.10"}, 60).SetVal(int64(2))

		h := HTTPRateLimitMiddlewareRedis(rdb, log, LimiterTypeIP, 10, 60)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "8", w.Header().Get("X-RateLimit-Remaining"))
		require.NoError(t, redMock.ExpectationsWereMet())
	})

	t.Run("blocks over limit with 429 and Retry-After", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		redMock.ExpectEvalSha(rateLimitScript.Hash(), []string{"rate_limit:ip:203.0.113.11"}, 60).SetVal(int64(6))

		h := HTTPRateLimitMiddlewareRedis(rdb, log, LimiterTypeIP, 5, 60)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "203.0.113.11:1234"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
		assert.Equal(t, "60", w.Header().Get("Retry-After"))
		require.NoError(t, redMock.ExpectationsWereMet())
	})

	t.Run("LimiterTypeUser uses user_id from context and falls back to IP", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		redMock.ExpectEvalSha(rateLimitScript.Hash(), []string{"rate_limit:user:user-42"}, 60).SetVal(int64(1))
		redMock.ExpectEvalSha(rateLimitScript.Hash(), []string{"rate_limit:ip_fallback:203.0.113.12"}, 60).SetVal(int64(1))

		h := HTTPRateLimitMiddlewareRedis(rdb, log, LimiterTypeUser, 100, 60)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "203.0.113.12:1234"
		req1 = req1.WithContext(delivery.SetContextValue(req1.Context(), delivery.UserIDKey, "user-42"))
		w1 := httptest.NewRecorder()
		h.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "203.0.113.12:1234"
		w2 := httptest.NewRecorder()
		h.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)

		require.NoError(t, redMock.ExpectationsWereMet())
	})

	t.Run("fails open on redis error", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		redMock.ExpectEvalSha(rateLimitScript.Hash(), []string{"rate_limit:ip:203.0.113.13"}, 60).SetErr(errors.New("redis down"))

		h := HTTPRateLimitMiddlewareRedis(rdb, log, LimiterTypeIP, 5, 60)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "203.0.113.13:1234"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		require.NoError(t, redMock.ExpectationsWereMet())
	})
}

func TestHTTPRateLimitMiddlewareMemory(t *testing.T) {
	h := HTTPRateLimitMiddlewareMemory(1, 2)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// 2 burst requests succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/mem", nil)
		req.RemoteAddr = "198.51.100.1:1000"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "2", w.Header().Get("X-RateLimit-Limit"))
	}

	// 3rd request blocked with 429 and headers
	req3 := httptest.NewRequest(http.MethodGet, "/mem", nil)
	req3.RemoteAddr = "198.51.100.1:1000"
	w3 := httptest.NewRecorder()
	h.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
	assert.Equal(t, "0", w3.Header().Get("X-RateLimit-Remaining"))
	assert.Equal(t, "1", w3.Header().Get("Retry-After"))

	// Different IP is isolated and allowed
	reqOther := httptest.NewRequest(http.MethodGet, "/mem", nil)
	reqOther.RemoteAddr = "198.51.100.2:1000"
	wOther := httptest.NewRecorder()
	h.ServeHTTP(wOther, reqOther)
	assert.Equal(t, http.StatusOK, wOther.Code)

	// Localhost is whitelisted
	reqLocal := httptest.NewRequest(http.MethodGet, "/mem", nil)
	reqLocal.RemoteAddr = "127.0.0.1:1000"
	wLocal := httptest.NewRecorder()
	h.ServeHTTP(wLocal, reqLocal)
	assert.Equal(t, http.StatusOK, wLocal.Code)
}

func TestHTTPTrustedProxyMiddleware(t *testing.T) {
	log := newDiscardLogger()
	mw := HTTPTrustedProxyMiddleware([]string{"10.0.0.0/8"}, log)

	var resolvedIP string
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resolvedIP = extractClientIP(r)
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("trusts X-Forwarded-For from trusted proxy", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.5:4321"
		req.Header.Set("X-Forwarded-For", "198.51.100.20, 10.0.0.2")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, "198.51.100.20", resolvedIP)
	})

	t.Run("ignores spoofed X-Forwarded-For from untrusted peer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "203.0.113.55:4321"
		req.Header.Set("X-Forwarded-For", "127.0.0.1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, "203.0.113.55", resolvedIP)
	})
}

func TestHTTPIdempotencyMiddleware(t *testing.T) {
	log := newDiscardLogger()

	t.Run("nil redis client passes through", func(t *testing.T) {
		h := HTTPIdempotencyMiddleware(nil, log, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))

		req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewBufferString(`{"a":1}`))
		req.Header.Set("Idempotency-Key", "k1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GET requests bypass idempotency", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		h := HTTPIdempotencyMiddleware(rdb, log, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/items", nil)
		req.Header.Set("Idempotency-Key", "k1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		require.NoError(t, redMock.ExpectationsWereMet())
	})

	t.Run("first POST caches 2xx response and restores body for handler", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		bodyStr := `{"name":"project-a"}`
		sum := sha256.Sum256([]byte(bodyStr))
		payloadHash := hex.EncodeToString(sum[:])

		respBody := `{"id":"p1"}`
		cached := IdempotencyResponse{
			Status:  http.StatusCreated,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    []byte(respBody),
		}
		cachedBytes, _ := json.Marshal(cached)

		redMock.ExpectSetNX("idempotency_lock:idem-1", payloadHash, 10*time.Second).SetVal(true)
		redMock.ExpectGet("idempotency:idem-1").RedisNil()
		redMock.ExpectSet("idempotency:idem-1", cachedBytes, time.Minute).SetVal("OK")
		redMock.ExpectDel("idempotency_lock:idem-1").SetVal(1)

		var handlerBody string
		h := HTTPIdempotencyMiddleware(rdb, log, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			handlerBody = string(b)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(respBody))
		}))

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(bodyStr))
		req.Header.Set("Idempotency-Key", "idem-1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, bodyStr, handlerBody)
		assert.Equal(t, respBody, w.Body.String())
		require.NoError(t, redMock.ExpectationsWereMet())
	})

	t.Run("replays cached response with X-Cache-Lookup header", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		bodyStr := `{"name":"project-a"}`
		sum := sha256.Sum256([]byte(bodyStr))
		payloadHash := hex.EncodeToString(sum[:])

		cached := IdempotencyResponse{
			Status:  http.StatusCreated,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    []byte(`{"id":"p1"}`),
		}
		cachedBytes, _ := json.Marshal(cached)

		redMock.ExpectSetNX("idempotency_lock:idem-2", payloadHash, 10*time.Second).SetVal(true)
		redMock.ExpectGet("idempotency:idem-2").SetVal(string(cachedBytes))
		redMock.ExpectDel("idempotency_lock:idem-2").SetVal(1)

		called := false
		h := HTTPIdempotencyMiddleware(rdb, log, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(bodyStr))
		req.Header.Set("Idempotency-Key", "idem-2")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.False(t, called)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "HIT - Idempotency Replay", w.Header().Get("X-Cache-Lookup"))
		assert.Equal(t, `{"id":"p1"}`, w.Body.String())
		require.NoError(t, redMock.ExpectationsWereMet())
	})

	t.Run("returns 409 Conflict on payload mismatch while locked", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		bodyStr := `{"name":"different-payload"}`
		sum := sha256.Sum256([]byte(bodyStr))
		payloadHash := hex.EncodeToString(sum[:])

		redMock.ExpectSetNX("idempotency_lock:idem-3", payloadHash, 10*time.Second).SetVal(false)
		redMock.ExpectGet("idempotency_lock:idem-3").SetVal("other-hash")

		h := HTTPIdempotencyMiddleware(rdb, log, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		}))

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(bodyStr))
		req.Header.Set("Idempotency-Key", "idem-3")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		require.NoError(t, redMock.ExpectationsWereMet())
	})
}

func TestHTTPAuthMiddleware(t *testing.T) {
	log := newDiscardLogger()

	t.Run("valid Bearer token injects context", func(t *testing.T) {
		mockAuthUC := authMocks.NewMockAuthUseCase(t)
		mockTicket := mocking.NewMockTicketManager(t)
		mw := NewAuthMiddleware(mockAuthUC, log, mockTicket)

		claims := &jwt.Claims{UserID: "u-1", SessionID: "s-1", Role: "role:user", Username: "alice"}
		mockAuthUC.On("ValidateAccessToken", "tok-1").Return(claims, nil)
		mockAuthUC.On("Verify", mock.Anything, "u-1", "s-1").Return(&authModel.Auth{ID: "s-1"}, nil)

		var gotUserID, gotMethod string
		h := mw.HTTPValidateToken()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotUserID, _ = delivery.GetContextString(r.Context(), delivery.UserIDKey)
			gotMethod, _ = delivery.GetContextString(r.Context(), delivery.AuthMethodKey)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.Header.Set("Authorization", "Bearer tok-1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "u-1", gotUserID)
		assert.Equal(t, "jwt", gotMethod)
	})

	t.Run("missing token returns 401", func(t *testing.T) {
		mockAuthUC := authMocks.NewMockAuthUseCase(t)
		mw := NewAuthMiddleware(mockAuthUC, log, mocking.NewMockTicketManager(t))

		h := mw.HTTPValidateToken()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("skips validation when already authenticated via API key", func(t *testing.T) {
		mockAuthUC := authMocks.NewMockAuthUseCase(t)
		mw := NewAuthMiddleware(mockAuthUC, log, mocking.NewMockTicketManager(t))

		h := mw.HTTPValidateToken()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		req = req.WithContext(delivery.SetContextValue(req.Context(), delivery.APIKeyIDKey, "key-1"))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("revoked session returns 401", func(t *testing.T) {
		mockAuthUC := authMocks.NewMockAuthUseCase(t)
		mw := NewAuthMiddleware(mockAuthUC, log, mocking.NewMockTicketManager(t))

		claims := &jwt.Claims{UserID: "u-1", SessionID: "s-revoked", Role: "role:user", Username: "alice"}
		mockAuthUC.On("ValidateAccessToken", "tok-rev").Return(claims, nil)
		mockAuthUC.On("Verify", mock.Anything, "u-1", "s-revoked").Return(nil, nil)

		h := mw.HTTPValidateToken()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: "tok-rev"})
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHTTPAPIKeyMiddleware(t *testing.T) {
	log := newDiscardLogger()

	t.Run("authenticates valid key, checks scopes, and blocks session-only routes", func(t *testing.T) {
		mockUC := new(apiKeyMocks.MockApiKeyUseCase)
		mockUserRepo := new(userMocks.MockUserRepository)
		mw := NewAPIKeyMiddleware(mockUC, mockUserRepo, log)

		identity := &apiKeyModel.ApiKeyIdentity{
			ApiKeyID:       "ak-1",
			UserID:         "u-1",
			OrganizationID: "org-1",
			Username:       "bot",
			Scopes:         []string{"project:view"},
		}
		mockUC.On("Authenticate", mock.Anything, "sk_live_ok").Return(identity, nil)

		// Scoped read passes
		readHandler := delivery.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), mw.HTTPAuthenticate(), mw.HTTPRequireScopes("project:view"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		req.Header.Set("X-API-Key", "sk_live_ok")
		w := httptest.NewRecorder()
		readHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Missing scope returns 403
		writeHandler := delivery.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}), mw.HTTPAuthenticate(), mw.HTTPRequireScopes("project:manage"))

		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/projects", nil)
		req2.Header.Set("X-API-Key", "sk_live_ok")
		w2 := httptest.NewRecorder()
		writeHandler.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusForbidden, w2.Code)

		// Auto-scope matches GET /api/v1/projects -> project:view
		autoHandler := delivery.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), mw.HTTPAuthenticate(), mw.HTTPRequireScopeAuto())

		req3 := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		req3.Header.Set("X-API-Key", "sk_live_ok")
		w3 := httptest.NewRecorder()
		autoHandler.ServeHTTP(w3, req3)
		assert.Equal(t, http.StatusOK, w3.Code)

		// RequireUserSession rejects API key with 403
		sessionOnly := delivery.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), mw.HTTPAuthenticate(), mw.HTTPRequireUserSession())

		req4 := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req4.Header.Set("X-API-Key", "sk_live_ok")
		w4 := httptest.NewRecorder()
		sessionOnly.ServeHTTP(w4, req4)
		assert.Equal(t, http.StatusForbidden, w4.Code)
	})

	t.Run("invalid API key returns 401", func(t *testing.T) {
		mockUC := new(apiKeyMocks.MockApiKeyUseCase)
		mw := NewAPIKeyMiddleware(mockUC, nil, log)
		mockUC.On("Authenticate", mock.Anything, "sk_live_bad").Return(nil, errors.New("invalid key"))

		h := mw.HTTPAuthenticate()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		req.Header.Set("X-API-Key", "sk_live_bad")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHTTPCasbinMiddleware(t *testing.T) {
	log := newDiscardLogger()

	t.Run("nil enforcer returns 403", func(t *testing.T) {
		h := HTTPCasbinMiddleware(nil, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("missing user returns 401", func(t *testing.T) {
		enforcer := new(httpMockCasbinEnforcer)
		h := HTTPCasbinMiddleware(enforcer, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("enforces policy with org domain and allows/denies", func(t *testing.T) {
		enforcer := new(httpMockCasbinEnforcer)
		enforcer.On("Enforce", "u-1", "org-9", "/api/v1/projects", "GET").Return(true, nil)
		enforcer.On("Enforce", "u-1", "org-9", "/api/v1/projects", "DELETE").Return(false, nil)

		h := HTTPCasbinMiddleware(enforcer, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ctx := delivery.SetContextValue(httptest.NewRequest(http.MethodGet, "/", nil).Context(), delivery.UserIDKey, "u-1")
		ctx = database.SetOrganizationContext(ctx, "org-9")

		reqAllow := httptest.NewRequest(http.MethodGet, "/api/v1/projects/", nil).WithContext(ctx)
		wAllow := httptest.NewRecorder()
		h.ServeHTTP(wAllow, reqAllow)
		assert.Equal(t, http.StatusOK, wAllow.Code)

		reqDeny := httptest.NewRequest(http.MethodDelete, "/api/v1/projects", nil).WithContext(ctx)
		wDeny := httptest.NewRecorder()
		h.ServeHTTP(wDeny, reqDeny)
		assert.Equal(t, http.StatusForbidden, wDeny.Code)

		enforcer.AssertExpectations(t)
	})
}

func TestHTTPTenantMiddleware(t *testing.T) {
	log := newDiscardLogger()

	t.Run("HTTPRequireOrganization validates membership and sets context", func(t *testing.T) {
		mockOrgRepo := orgMocks.NewMockOrganizationRepository(t)
		mockReader := orgMocks.NewMockIOrganizationReader(t)
		mw := NewTenantMiddleware(mockOrgRepo, mockReader, log)

		mockReader.On("ValidateMembership", mock.Anything, "org-1", "u-1").Return(true, nil)

		var gotOrg string
		h := mw.HTTPRequireOrganization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotOrg = database.GetOrganizationID(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		req.Header.Set(OrgIDHeader, "org-1")
		req = req.WithContext(delivery.SetContextValue(req.Context(), delivery.UserIDKey, "u-1"))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "org-1", gotOrg)
	})

	t.Run("HTTPRequireOrganization rejects missing org and non-members", func(t *testing.T) {
		mockOrgRepo := orgMocks.NewMockOrganizationRepository(t)
		mockReader := orgMocks.NewMockIOrganizationReader(t)
		mw := NewTenantMiddleware(mockOrgRepo, mockReader, log)

		mockReader.On("ValidateMembership", mock.Anything, "org-2", "u-1").Return(false, nil)

		h := mw.HTTPRequireOrganization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Missing org -> 400
		reqMissing := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		reqMissing = reqMissing.WithContext(delivery.SetContextValue(reqMissing.Context(), delivery.UserIDKey, "u-1"))
		wMissing := httptest.NewRecorder()
		h.ServeHTTP(wMissing, reqMissing)
		assert.Equal(t, http.StatusBadRequest, wMissing.Code)

		// Non-member -> 403
		reqDeny := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		reqDeny.Header.Set(OrgIDHeader, "org-2")
		reqDeny = reqDeny.WithContext(delivery.SetContextValue(reqDeny.Context(), delivery.UserIDKey, "u-1"))
		wDeny := httptest.NewRecorder()
		h.ServeHTTP(wDeny, reqDeny)
		assert.Equal(t, http.StatusForbidden, wDeny.Code)
	})

	t.Run("HTTPOptionalOrganization resolves slug when present", func(t *testing.T) {
		mockOrgRepo := orgMocks.NewMockOrganizationRepository(t)
		mockReader := orgMocks.NewMockIOrganizationReader(t)
		mw := NewTenantMiddleware(mockOrgRepo, mockReader, log)

		mockOrgRepo.On("FindBySlug", mock.Anything, "acme").Return(&orgEntity.Organization{ID: "org-acme"}, nil)

		var gotOrg string
		h := mw.HTTPOptionalOrganization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotOrg = database.GetOrganizationID(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		req.Header.Set(OrgSlugHeader, "acme")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "org-acme", gotOrg)
	})
}

func TestHTTPUserStatusMiddleware(t *testing.T) {
	log := newDiscardLogger()

	t.Run("active passes, suspended and banned return 403", func(t *testing.T) {
		mockRepo := new(userMocks.MockUserRepository)
		mockRepo.On("FindByID", mock.Anything, "u-active").Return(&userEntity.User{ID: "u-active", Status: userEntity.UserStatusActive}, nil)
		mockRepo.On("FindByID", mock.Anything, "u-susp").Return(&userEntity.User{ID: "u-susp", Status: userEntity.UserStatusSuspended}, nil)
		mockRepo.On("FindByID", mock.Anything, "u-ban").Return(&userEntity.User{ID: "u-ban", Status: userEntity.UserStatusBanned}, nil)

		h := HTTPUserStatusMiddleware(mockRepo, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		for _, tc := range []struct {
			uid  string
			code int
		}{
			{"u-active", http.StatusOK},
			{"u-susp", http.StatusForbidden},
			{"u-ban", http.StatusForbidden},
		} {
			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			req = req.WithContext(delivery.SetContextValue(req.Context(), delivery.UserIDKey, tc.uid))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			assert.Equal(t, tc.code, w.Code)
		}
	})

	t.Run("missing user context returns 401", func(t *testing.T) {
		mockRepo := new(userMocks.MockUserRepository)
		h := HTTPUserStatusMiddleware(mockRepo, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHTTPValidateWebSocketToken(t *testing.T) {
	log := newDiscardLogger()

	t.Run("missing ticket returns 401", func(t *testing.T) {
		mw := NewAuthMiddleware(authMocks.NewMockAuthUseCase(t), log, mocking.NewMockTicketManager(t))
		h := mw.HTTPValidateWebSocketToken()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ws", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid ticket injects user and org context", func(t *testing.T) {
		mockTicket := mocking.NewMockTicketManager(t)
		mockTicket.On("ValidateTicket", mock.Anything, "ticket-ok").Return(&ws.UserContext{
			UserID:         "u-ws",
			SessionID:      "s-ws",
			Role:           "role:user",
			Username:       "wsuser",
			OrganizationID: "org-ws",
		}, nil)

		mw := NewAuthMiddleware(authMocks.NewMockAuthUseCase(t), log, mockTicket)

		var gotUser, gotOrg string
		h := mw.HTTPValidateWebSocketToken()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotUser, _ = delivery.GetContextString(r.Context(), delivery.UserIDKey)
			gotOrg = database.GetOrganizationID(r.Context())
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ws?ticket=ticket-ok", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "u-ws", gotUser)
		assert.Equal(t, "org-ws", gotOrg)
	})

	t.Run("invalid ticket returns 401", func(t *testing.T) {
		mockTicket := mocking.NewMockTicketManager(t)
		mockTicket.On("ValidateTicket", mock.Anything, "bad-ticket").Return(nil, errors.New("invalid ticket"))

		mw := NewAuthMiddleware(authMocks.NewMockAuthUseCase(t), log, mockTicket)
		h := mw.HTTPValidateWebSocketToken()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/ws?ticket=bad-ticket", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestScopeFromMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected string
	}{
		{http.MethodGet, "view"},
		{http.MethodHead, "view"},
		{http.MethodPost, "create"},
		{http.MethodPut, "update"},
		{http.MethodPatch, "update"},
		{http.MethodDelete, "delete"},
		{http.MethodOptions, "view"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			assert.Equal(t, tt.expected, ScopeFromMethod(tt.method))
		})
	}
}

func TestRequiredScopeFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		method   string
		expected string
		ok       bool
	}{
		{name: "nested me path", path: "/api/v1/users/me", method: http.MethodGet, expected: "user:view", ok: true},
		{name: "nested avatar path", path: "/api/v1/users/me/avatar", method: http.MethodPatch, expected: "user:update", ok: true},
		{name: "organization scope alias", path: "/api/v1/organizations/org-123", method: http.MethodGet, expected: "org:view", ok: true},
		{name: "trailing slash path", path: "/api/v1/projects/", method: http.MethodGet, expected: "project:view", ok: true},
		{name: "invalid short path", path: "/health", method: http.MethodGet, expected: "", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, ok := requiredScopeFromRequest(tt.path, tt.method)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.expected, scope)
		})
	}
}

func TestHasRequiredScope(t *testing.T) {
	assert.True(t, hasRequiredScope([]string{"*"}, "anything:here"))
	assert.True(t, hasRequiredScope([]string{"project:*"}, "project:manage"))
	assert.True(t, hasRequiredScope([]string{"project:manage"}, "project:delete"))
	assert.True(t, hasRequiredScope([]string{"project:read"}, "project:view"))
	assert.True(t, hasRequiredScope([]string{"project:view"}, "project:read"))
	assert.True(t, hasRequiredScope([]string{"project:view"}))
	assert.False(t, hasRequiredScope([]string{"project:view"}, "project:create"))
	assert.False(t, hasRequiredScope([]string{"invalid-format"}, "project:view"))
}

func TestHTTPMiddlewareEdgeCases(t *testing.T) {
	log := newDiscardLogger()

	t.Run("HTTPAuthMiddleware invalid header format, signature error, and verify DB error", func(t *testing.T) {
		mockAuthUC := authMocks.NewMockAuthUseCase(t)
		mw := NewAuthMiddleware(mockAuthUC, log, mocking.NewMockTicketManager(t))
		h := mw.HTTPValidateToken()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Invalid Authorization header format -> 401
		reqFmt := httptest.NewRequest(http.MethodGet, "/me", nil)
		reqFmt.Header.Set("Authorization", "InvalidToken")
		wFmt := httptest.NewRecorder()
		h.ServeHTTP(wFmt, reqFmt)
		assert.Equal(t, http.StatusUnauthorized, wFmt.Code)

		// Invalid signature -> 401
		mockAuthUC.On("ValidateAccessToken", "bad-sig").Return(nil, errors.New("invalid signature"))
		reqSig := httptest.NewRequest(http.MethodGet, "/me", nil)
		reqSig.Header.Set("Authorization", "Bearer bad-sig")
		wSig := httptest.NewRecorder()
		h.ServeHTTP(wSig, reqSig)
		assert.Equal(t, http.StatusUnauthorized, wSig.Code)

		// Verify DB error -> 500
		claims := &jwt.Claims{UserID: "u-err", SessionID: "s-err", Role: "role:user", Username: "bob"}
		mockAuthUC.On("ValidateAccessToken", "tok-dberr").Return(claims, nil)
		mockAuthUC.On("Verify", mock.Anything, "u-err", "s-err").Return(nil, errors.New("db down"))
		reqDB := httptest.NewRequest(http.MethodGet, "/me", nil)
		reqDB.Header.Set("Authorization", "Bearer tok-dberr")
		wDB := httptest.NewRecorder()
		h.ServeHTTP(wDB, reqDB)
		assert.Equal(t, http.StatusInternalServerError, wDB.Code)
	})

	t.Run("HTTPAPIKeyMiddleware JWT bypass, RequireScopeAuto short path, and wildcard scope", func(t *testing.T) {
		mockUC := new(apiKeyMocks.MockApiKeyUseCase)
		mw := NewAPIKeyMiddleware(mockUC, nil, log)

		// JWT auth bypasses RequireScopes and RequireUserSession
		hJWT := delivery.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}), mw.HTTPRequireScopes("admin:manage"), mw.HTTPRequireUserSession())

		reqJWT := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		reqJWT = reqJWT.WithContext(delivery.SetContextValue(reqJWT.Context(), delivery.AuthMethodKey, "jwt"))
		wJWT := httptest.NewRecorder()
		hJWT.ServeHTTP(wJWT, reqJWT)
		assert.Equal(t, http.StatusOK, wJWT.Code)

		// Global wildcard allows arbitrary scope
		hWildcard := mw.HTTPRequireScopes("anything:here")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		reqWild := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctxWild := delivery.SetContextValue(reqWild.Context(), delivery.AuthMethodKey, "api_key")
		ctxWild = delivery.SetContextValue(ctxWild, delivery.ScopesKey, []string{"*"})
		wWild := httptest.NewRecorder()
		hWildcard.ServeHTTP(wWild, reqWild.WithContext(ctxWild))
		assert.Equal(t, http.StatusOK, wWild.Code)

		// RequireScopeAuto denies undeterminable short path
		hAuto := mw.HTTPRequireScopeAuto()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		reqShort := httptest.NewRequest(http.MethodGet, "/health", nil)
		wShort := httptest.NewRecorder()
		hAuto.ServeHTTP(wShort, reqShort.WithContext(ctxWild))
		assert.Equal(t, http.StatusForbidden, wShort.Code)
	})

	t.Run("HTTPCasbinMiddleware global domain fallback and enforcer error", func(t *testing.T) {
		enforcer := new(httpMockCasbinEnforcer)
		enforcer.On("Enforce", "u-1", "global", "/api/v1/users", "GET").Return(true, nil)
		enforcer.On("Enforce", "u-2", "global", "/api/v1/users", "GET").Return(false, errors.New("casbin error"))

		h := HTTPCasbinMiddleware(enforcer, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		reqOK := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		reqOK = reqOK.WithContext(delivery.SetContextValue(reqOK.Context(), delivery.UserIDKey, "u-1"))
		wOK := httptest.NewRecorder()
		h.ServeHTTP(wOK, reqOK)
		assert.Equal(t, http.StatusOK, wOK.Code)

		reqErr := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		reqErr = reqErr.WithContext(delivery.SetContextValue(reqErr.Context(), delivery.UserIDKey, "u-2"))
		wErr := httptest.NewRecorder()
		h.ServeHTTP(wErr, reqErr)
		assert.Equal(t, http.StatusInternalServerError, wErr.Code)

		enforcer.AssertExpectations(t)
	})

	t.Run("HTTPTenantMiddleware slug lookup, path param, unauthenticated, error, and cache invalidation", func(t *testing.T) {
		mockOrgRepo := orgMocks.NewMockOrganizationRepository(t)
		mockReader := orgMocks.NewMockIOrganizationReader(t)
		mw := NewTenantMiddleware(mockOrgRepo, mockReader, log)

		mockOrgRepo.On("FindBySlug", mock.Anything, "my-org").Return(&orgEntity.Organization{ID: "org-slug-1", Slug: "my-org"}, nil)
		mockOrgRepo.On("FindBySlug", mock.Anything, "missing-org").Return(nil, nil)
		mockReader.On("ValidateMembership", mock.Anything, "org-slug-1", "u-1").Return(true, nil)
		mockReader.On("ValidateMembership", mock.Anything, "org-path-1", "u-1").Return(true, nil)
		mockReader.On("ValidateMembership", mock.Anything, "org-err", "u-1").Return(false, errors.New("redis down"))
		mockReader.On("InvalidateMembershipCache", mock.Anything, "org-1", "u-1").Return(nil)

		h := mw.HTTPRequireOrganization()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// Unauthenticated -> 401
		reqNoAuth := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		reqNoAuth.Header.Set(OrgIDHeader, "org-1")
		wNoAuth := httptest.NewRecorder()
		h.ServeHTTP(wNoAuth, reqNoAuth)
		assert.Equal(t, http.StatusUnauthorized, wNoAuth.Code)

		// Slug lookup -> 200
		reqSlug := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		reqSlug.Header.Set(OrgSlugHeader, "my-org")
		reqSlug = reqSlug.WithContext(delivery.SetContextValue(reqSlug.Context(), delivery.UserIDKey, "u-1"))
		wSlug := httptest.NewRecorder()
		h.ServeHTTP(wSlug, reqSlug)
		assert.Equal(t, http.StatusOK, wSlug.Code)

		// Missing slug -> 404
		reqNotFound := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		reqNotFound.Header.Set(OrgSlugHeader, "missing-org")
		reqNotFound = reqNotFound.WithContext(delivery.SetContextValue(reqNotFound.Context(), delivery.UserIDKey, "u-1"))
		wNotFound := httptest.NewRecorder()
		h.ServeHTTP(wNotFound, reqNotFound)
		assert.Equal(t, http.StatusNotFound, wNotFound.Code)

		// PathValue("id") -> 200
		reqPath := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/org-path-1/presence", nil)
		reqPath.SetPathValue("id", "org-path-1")
		reqPath = reqPath.WithContext(delivery.SetContextValue(reqPath.Context(), delivery.UserIDKey, "u-1"))
		wPath := httptest.NewRecorder()
		h.ServeHTTP(wPath, reqPath)
		assert.Equal(t, http.StatusOK, wPath.Code)

		// Reader error -> 500
		reqErr := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		reqErr.Header.Set(OrgIDHeader, "org-err")
		reqErr = reqErr.WithContext(delivery.SetContextValue(reqErr.Context(), delivery.UserIDKey, "u-1"))
		wErr := httptest.NewRecorder()
		h.ServeHTTP(wErr, reqErr)
		assert.Equal(t, http.StatusInternalServerError, wErr.Code)

		// InvalidateMembershipCache
		require.NoError(t, mw.InvalidateMembershipCache(httptest.NewRequest(http.MethodGet, "/", nil).Context(), "org-1", "u-1"))
	})

	t.Run("HTTPUserStatusMiddleware DB error returns 500", func(t *testing.T) {
		mockRepo := new(userMocks.MockUserRepository)
		mockRepo.On("FindByID", mock.Anything, "u-err").Return(nil, errors.New("db error"))

		h := HTTPUserStatusMiddleware(mockRepo, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		req = req.WithContext(delivery.SetContextValue(req.Context(), delivery.UserIDKey, "u-err"))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("HTTPCORSMiddleware empty origins omits CORS headers", func(t *testing.T) {
		h := HTTPCORSMiddleware([]string{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		req.Header.Set("Origin", "https://any-origin.example.com")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("HTTPRateLimitMiddlewareRedis exact limit boundary passes", func(t *testing.T) {
		rdb, redMock := redismock.NewClientMock()
		redMock.ExpectEvalSha(rateLimitScript.Hash(), []string{"rate_limit:ip:203.0.113.99"}, 60).SetVal(int64(10))

		h := HTTPRateLimitMiddlewareRedis(rdb, log, LimiterTypeIP, 10, 60)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "203.0.113.99:1234"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
		require.NoError(t, redMock.ExpectationsWereMet())
	})
}
