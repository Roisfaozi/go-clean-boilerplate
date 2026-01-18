package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupCORSTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ============================================================================
// ✅ BASIC FUNCTIONALITY TESTS
// ============================================================================

func TestCORSMiddleware_CanBeInstantiated(t *testing.T) {
	// Test that middleware can be created without panic
	middleware := CORSMiddleware([]string{})
	assert.NotNil(t, middleware)
}

func TestCORSMiddleware_WithEmptyOrigins(t *testing.T) {
	router := setupCORSTest()
	
	// Setup CORS with empty origins (defaults to wildcard)
	router.Use(CORSMiddleware([]string{}))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test request with origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Request should succeed
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestCORSMiddleware_WithSpecificOrigins(t *testing.T) {
	router := setupCORSTest()
	
	// Setup CORS with specific origins
	origins := []string{"https://example.com", "https://app.example.com"}
	router.Use(CORSMiddleware(origins))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Request should be processed (CORS enforcement is browser-side)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Request without Origin header should succeed
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORSMiddleware_RequestWithoutOrigin(t *testing.T) {
	router := setupCORSTest()
	
	router.Use(CORSMiddleware([]string{}))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Same-origin request (no Origin header)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Should succeed
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORSMiddleware_OptionsRequest(t *testing.T) {
	router := setupCORSTest()
	
	// Setup CORS middleware
	router.Use(CORSMiddleware([]string{}))
	
	// Register the actual route handler  
	router.POST("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// OPTIONS request should be handled by CORS middleware
	req := httptest.NewRequest("OPTIONS", "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	
	// Should not panic
	assert.NotPanics(t, func() {
		router.ServeHTTP(w, req)
	})
	
	// Should return some valid response (not 500)
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestCORSMiddleware_AllowedMethods(t *testing.T) {
	router := setupCORSTest()
	
	router.Use(CORSMiddleware([]string{}))
	
	// Register multiple method handlers
	router.GET("/resource", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/resource", func(c *gin.Context) { c.Status(http.StatusCreated) })
	router.DELETE("/resource", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	tests := []struct {
		method       string
		expectedCode int
	}{
		{"GET", http.StatusOK},
		{"POST", http.StatusCreated},
		{"DELETE", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/resource", nil)
			req.Header.Set("Origin", "https://example.com")
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
