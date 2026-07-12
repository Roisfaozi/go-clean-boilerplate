package tus

import (
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
)

func TestSwaggerDoc(t *testing.T) {
    gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
    ctrl := &TusController{}

	t.Run("Positive - TusUpload does not panic", func(t *testing.T) {
		ctrl.TusUpload(c)
	})
	t.Run("Positive - TusPatch does not panic", func(t *testing.T) {
		ctrl.TusPatch(c)
	})
	t.Run("Positive - TusHead does not panic", func(t *testing.T) {
		ctrl.TusHead(c)
	})
	t.Run("Positive - TusDelete does not panic", func(t *testing.T) {
		ctrl.TusDelete(c)
	})
}
