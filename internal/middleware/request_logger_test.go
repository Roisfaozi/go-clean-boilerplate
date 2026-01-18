package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func setupRequestLoggerTest() (*gin.Engine, *logrus.Logger, *bytes.Buffer) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	log := logrus.New()
	logBuffer := &bytes.Buffer{}
	log.SetOutput(logBuffer)
	log.SetFormatter(&logrus.JSONFormatter{})
	
	return router, log, logBuffer
}

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestRequestLogger_LogsRequest(t *testing.T) {
	router, log, logBuffer := setupRequestLoggerTest()
	
	router.Use(RequestLogger(log))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Parse log output
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "http_request")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "200")
	assert.Contains(t, logOutput, "TestAgent/1.0")
}

func TestRequestLogger_LogsWithRequestID(t *testing.T) {
	router, log, logBuffer := setupRequestLoggerTest()
	
	// Add RequestID middleware first
	router.Use(RequestIDMiddleware())
	router.Use(RequestLogger(log))
	
	router.GET("/test-with-id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test-with-id", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check that request_id is in response header
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
	
	// Log should contain the request
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "http_request")
	assert.Contains(t, logOutput, "/test-with-id")
}

// ❌ NEGATIVE CASES

func TestRequestLogger_LogsClientError(t *testing.T) {
	router, log, logBuffer := setupRequestLoggerTest()
	
	router.Use(RequestLogger(log))
	
	router.GET("/not-found", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	req := httptest.NewRequest("GET", "/not-found", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(logBuffer.Bytes(), &logEntry)
	assert.NoError(t, err)
	
	assert.Equal(t, "http_request", logEntry["type"])
	assert.Equal(t, float64(404), logEntry["status"])
	assert.Equal(t, "warning", logEntry["level"]) // 4xx should be warning
}

func TestRequestLogger_LogsServerError(t *testing.T) {
	router, log, logBuffer := setupRequestLoggerTest()
	
	router.Use(RequestLogger(log))
	
	router.GET("/server-error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})

	req := httptest.NewRequest("GET", "/server-error", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(logBuffer.Bytes(), &logEntry)
	assert.NoError(t, err)
	
	assert.Equal(t, "http_request", logEntry["type"])
	assert.Equal(t, float64(500), logEntry["status"])
	assert.Equal(t, "error", logEntry["level"]) // 5xx should be error
}

// 🔄 EDGE CASES

func TestRequestLogger_LogsLatency(t *testing.T) {
	router, log, logBuffer := setupRequestLoggerTest()
	
	router.Use(RequestLogger(log))
	
	router.GET("/slow", func(c *gin.Context) {
		// Simulate some processing
		c.JSON(http.StatusOK, gin.H{"message": "done"})
	})

	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(logBuffer.Bytes(), &logEntry)
	assert.NoError(t, err)
	
	// Should have latency fields
	assert.Contains(t, logEntry, "latency_ns")
	assert.Contains(t, logEntry, "latency_ms")
	assert.Greater(t, logEntry["latency_ns"], float64(0))
}

func TestRequestLogger_LogsDataLength(t *testing.T) {
	router, log, logBuffer := setupRequestLoggerTest()
	
	router.Use(RequestLogger(log))
	
	router.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    []int{1, 2, 3, 4, 5},
		})
	})

	req := httptest.NewRequest("GET", "/data", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Parse log output
	var logEntry map[string]interface{}
	err := json.Unmarshal(logBuffer.Bytes(), &logEntry)
	assert.NoError(t, err)
	// Should have data_length field
	assert.Contains(t, logEntry, "data_length")
	assert.Greater(t, logEntry["data_length"], float64(0))
}
