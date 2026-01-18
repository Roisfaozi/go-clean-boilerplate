package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setupRecoveryTest() (*gin.Engine, *logrus.Logger) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	log := logrus.New()
	log.SetOutput(io.Discard) // Suppress logs in tests
	
	return router, log
}

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestRecoveryMiddleware_PanicRecovery(t *testing.T) {
	router, log := setupRecoveryTest()
	
	router.Use(RecoveryMiddleware(log))
	
	// Route that panics
	router.GET("/panic", func(c *gin.Context) {
		panic("something went wrong!")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	req.Header.Set("X-Request-ID", "test-request-123")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert - Should recover and return 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal server error")
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	router, log := setupRecoveryTest()
	
	router.Use(RecoveryMiddleware(log))
	
	// Normal route that doesn't panic
	router.GET("/normal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/normal", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert - Should work normally
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// 🔄 EDGE CASES

func TestRecoveryMiddleware_PanicWithRequestID(t *testing.T) {
	router, log := setupRecoveryTest()
	
	router.Use(RecoveryMiddleware(log))
	
	// Route that panics with request_id in context
	router.GET("/panic-with-id", func(c *gin.Context) {
		c.Set("request_id", "ctx-request-456")
		panic("panic with context request ID")
	})

	req := httptest.NewRequest("GET", "/panic-with-id", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert - Should recover and use context request_id
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecoveryMiddleware_PanicNilError(t *testing.T) {
	router, log := setupRecoveryTest()
	
	router.Use(RecoveryMiddleware(log))
	
	// Route that panics with nil
	router.GET("/panic-nil", func(c *gin.Context) {
		panic(nil)
	})

	req := httptest.NewRequest("GET", "/panic-nil", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert - Should still recover
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecoveryMiddleware_PanicDifferentTypes(t *testing.T) {
	router, log := setupRecoveryTest()
	
	router.Use(RecoveryMiddleware(log))
	
	testCases := []struct {
		name      string
		panicVal  interface{}
		path      string
	}{
		{"String panic", "error string", "/panic-string"},
		{"Int panic", 42, "/panic-int"},
		{"Struct panic", struct{ Msg string }{"error"}, "/panic-struct"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router.GET(tc.path, func(c *gin.Context) {
				panic(tc.panicVal)
			})

			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)

			// Assert - All should be recovered
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	}
}
