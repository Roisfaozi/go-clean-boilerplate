package tus

import (
	"testing"
	"net/http/httptest"
	"github.com/gin-gonic/gin"
)

func TestSwaggerDoc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ctrl := &TusController{}
	ctrl.TusUpload(c)
	ctrl.TusPatch(c)
	ctrl.TusHead(c)
	ctrl.TusDelete(c)
}
