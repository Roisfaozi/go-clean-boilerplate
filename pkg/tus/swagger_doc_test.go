package tus

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSwaggerStubs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	ctrl := &TusController{}

	// Call stubs to satisfy coverage
	ctrl.TusUpload(c)
	ctrl.TusPatch(c)
	ctrl.TusHead(c)
	ctrl.TusDelete(c)
}
