package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInternalServerError_DebugMode(t *testing.T) {
	gin.SetMode(gin.DebugMode)
	defer gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := errors.New("database connection failed")
	response.InternalServerError(c, err, "Something wrong")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.WebResponseError[any]
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, "database connection failed", resp.Error)
	assert.Equal(t, "Something wrong", resp.Message)
}

func TestInternalServerError_ReleaseMode(t *testing.T) {

	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := errors.New("sensitive db error")
	response.InternalServerError(c, err, "Something wrong")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.WebResponseError[any]
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Equal(t, "Internal Server Error", resp.Error)
	assert.Equal(t, "Something wrong", resp.Message)
}

func TestHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{"BadRequest", exception.ErrBadRequest, http.StatusBadRequest},
		{"Unauthorized", exception.ErrUnauthorized, http.StatusUnauthorized},
		{"Forbidden", exception.ErrForbidden, http.StatusForbidden},
		{"NotFound", exception.ErrNotFound, http.StatusNotFound},
		{"Conflict", exception.ErrConflict, http.StatusConflict},
		{"ValidationError", exception.ErrValidationError, http.StatusUnprocessableEntity},
		{"UnprocessableEntity", exception.ErrUnprocessableEntity, http.StatusUnprocessableEntity},
		{"TooManyRequests", exception.ErrTooManyRequests, http.StatusTooManyRequests},
		{"UnknownError", errors.New("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			response.HandleError(c, tt.err, "error message")

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"foo": "bar"}
	response.Success(c, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.WebResponseSuccess[map[string]string]
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "bar", resp.Data["foo"])
}

func TestCreatedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"id": "123"}
	response.Created(c, data)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSuccessResponseWithPaging(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := []string{"item1", "item2"}
	paging := &response.PageMetadata{
		Page:      1,
		Limit:     10,
		TotalItem: 2,
		TotalPage: 1,
	}
	response.SuccessResponseWithPaging(c, data, paging)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.WebResponseSuccess[[]string]
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, int64(2), resp.Paging.TotalItem)
}
