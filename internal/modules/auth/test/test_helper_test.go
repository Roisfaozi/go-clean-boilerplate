package test_test

import (
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// Helper to create context with recorder
func newGinContextWithRecorder(w *httptest.ResponseRecorder) (*gin.Context, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	ctx, engine := gin.CreateTestContext(w)
	return ctx, engine
}
