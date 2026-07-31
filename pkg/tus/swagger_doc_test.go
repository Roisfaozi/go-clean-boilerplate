package tus

import (
	"testing"
	"github.com/gin-gonic/gin"
)

// TestSwaggerDocStubs_Positive verifies that the swagger doc stub methods do not panic
// and can be successfully invoked by the framework router.
func TestSwaggerDocStubs_Positive(t *testing.T) {
	c := &gin.Context{}
	ctrl := &TusController{}
	ctrl.TusUpload(c)
	ctrl.TusPatch(c)
	ctrl.TusHead(c)
	ctrl.TusDelete(c)
}
