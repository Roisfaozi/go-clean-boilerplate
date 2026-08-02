package tus

import (
	"github.com/gin-gonic/gin"
	"testing"
)

func TestSwaggerDocStubs(t *testing.T) {
	ctrl := &TusController{}
	ctx := &gin.Context{}

	// They shouldn't panic
	ctrl.TusUpload(ctx)
	ctrl.TusPatch(ctx)
	ctrl.TusHead(ctx)
	ctrl.TusDelete(ctx)
}
