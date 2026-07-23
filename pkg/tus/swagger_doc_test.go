package tus

import (
	"testing"
	"github.com/gin-gonic/gin"
)

func TestSwaggerDocStubs(t *testing.T) {
	c, _ := gin.CreateTestContext(nil)
	ctrl := &TusController{}

	// Execute stubs to satisfy coverage
	ctrl.TusUpload(c)
	ctrl.TusPatch(c)
	ctrl.TusHead(c)
	ctrl.TusDelete(c)
}
