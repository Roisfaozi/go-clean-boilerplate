package response_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInternalServerError_Leakage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	sensitiveErr := errors.New("connection to 192.168.1.1:5432 failed: access denied")
	response.InternalServerError(c, sensitiveErr, "Something went wrong")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// SECURITY VERIFICATION: Ensure sensitive info is NOT leaked
	assert.NotContains(t, w.Body.String(), "192.168.1.1")
	assert.Contains(t, w.Body.String(), "Internal Server Error")
}
