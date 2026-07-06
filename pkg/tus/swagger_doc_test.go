package tus

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSwaggerDocMethods_AreRegisteredProperly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Since these are just swagger doc functions without implementation,
	// we test that they don't panic when invoked with a standard gin context
	// and they don't accidentally modify the response state (as they are stubs).

	tests := []struct {
		name   string
		method func(ctrl *TusController, c *gin.Context)
	}{
		{
			name:   "TusUpload stub executes safely",
			method: func(ctrl *TusController, c *gin.Context) { ctrl.TusUpload(c) },
		},
		{
			name:   "TusPatch stub executes safely",
			method: func(ctrl *TusController, c *gin.Context) { ctrl.TusPatch(c) },
		},
		{
			name:   "TusHead stub executes safely",
			method: func(ctrl *TusController, c *gin.Context) { ctrl.TusHead(c) },
		},
		{
			name:   "TusDelete stub executes safely",
			method: func(ctrl *TusController, c *gin.Context) { ctrl.TusDelete(c) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			ctrl := &TusController{}

			// Execute the swagger stub
			tt.method(ctrl, c)

			// Validate that the stub is truly a no-op (has no side effects on the context)
			assert.Equal(t, http.StatusOK, w.Code) // Default starting code
			assert.Empty(t, w.Body.String())
		})
	}
}
