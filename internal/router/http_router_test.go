package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStdHealth(t *testing.T) {
	healthFunc := GetStdHealth(nil, nil)

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	healthFunc(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"OK"`)
}

func TestBasicAuthHandler(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("metrics-data"))
	})

	authHandler := BasicAuthHandler(nextHandler, "admin", "secret123")

	t.Run("Reject without auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Reject invalid credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		req.SetBasicAuth("admin", "wrong")
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Allow valid credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		req.SetBasicAuth("admin", "secret123")
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "metrics-data", w.Body.String())
	})
}
