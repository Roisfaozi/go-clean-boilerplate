package tus

import (
	"testing"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
)

func TestSwaggerDoc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ctrl := &TusController{}
	_ = c
	ctrl.TusUpload(c)
	ctrl.TusPatch(c)
	ctrl.TusHead(c)
	ctrl.TusDelete(c)
}
