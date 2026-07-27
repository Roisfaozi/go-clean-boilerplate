package tus

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSwaggerDocStubs(t *testing.T) {
	c := &gin.Context{}
	ctrl := &TusController{}

	ctrl.TusUpload(c)
	ctrl.TusPatch(c)
	ctrl.TusHead(c)
	ctrl.TusDelete(c)
}
