package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	accessHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAccessTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func newTestAccessController(mockUseCase *mocks.MockIAccessUseCase) *accessHandler.AccessController {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	return accessHandler.NewAccessController(mockUseCase, v, log)
}

func TestAccessHandler_CreateAccessRight_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights", handler.CreateAccessRight)

	reqBody := model.CreateAccessRightRequest{
		Name:        "test_access_right",
		Description: "A description",
	}
	resBody := &model.AccessRightResponse{
		ID:   "1",
		Name: "test_access_right",
	}

	mockUseCase.On("CreateAccessRight", mock.Anything, reqBody).Return(resBody, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/access-rights", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_CreateAccessRight_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights", handler.CreateAccessRight)

	req, _ := http.NewRequest(http.MethodPost, "/access-rights", bytes.NewBufferString(`{"name":`)) // Malformed JSON
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "CreateAccessRight", mock.Anything, mock.Anything)
}

func TestAccessHandler_CreateAccessRight_ValidationErrors(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights", handler.CreateAccessRight)

	reqBody := model.CreateAccessRightRequest{
		Name: "",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/access-rights", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "CreateAccessRight", mock.Anything, mock.Anything)
}

func TestAccessHandler_CreateAccessRight_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights", handler.CreateAccessRight)

	reqBody := model.CreateAccessRightRequest{
		Name: "test_access_right",
	}
	mockUseCase.On("CreateAccessRight", mock.Anything, reqBody).Return(nil, errors.New("db error"))

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/access-rights", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_GetAllAccessRights_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.GET("/access-rights", handler.GetAllAccessRights)

	resBody := &model.AccessRightListResponse{
		Data: []model.AccessRightResponse{
			{ID: "1", Name: "right1"},
			{ID: "2", Name: "right2"},
		},
	}
	mockUseCase.On("GetAllAccessRights", mock.Anything).Return(resBody, nil)

	req, _ := http.NewRequest(http.MethodGet, "/access-rights", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_GetAllAccessRights_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.GET("/access-rights", handler.GetAllAccessRights)

	mockUseCase.On("GetAllAccessRights", mock.Anything).Return(nil, errors.New("db error"))

	req, _ := http.NewRequest(http.MethodGet, "/access-rights", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_CreateEndpoint_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/endpoints", handler.CreateEndpoint)

	reqBody := model.CreateEndpointRequest{
		Path:   "/test/path",
		Method: "GET",
	}
	resBody := &model.EndpointResponse{
		ID:     "1",
		Path:   "/test/path",
		Method: "GET",
	}

	mockUseCase.On("CreateEndpoint", mock.Anything, reqBody).Return(resBody, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/endpoints", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_CreateEndpoint_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/endpoints", handler.CreateEndpoint)

	req, _ := http.NewRequest(http.MethodPost, "/endpoints", bytes.NewBufferString(`{"path":`)) // Malformed JSON
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "CreateEndpoint", mock.Anything, mock.Anything)
}

func TestAccessHandler_CreateEndpoint_ValidationErrors(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/endpoints", handler.CreateEndpoint)

	reqBody := model.CreateEndpointRequest{
		Path: "",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/endpoints", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "CreateEndpoint", mock.Anything, mock.Anything)
}

func TestAccessHandler_CreateEndpoint_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/endpoints", handler.CreateEndpoint)

	reqBody := model.CreateEndpointRequest{
		Path:   "/test/path",
		Method: "GET",
	}
	mockUseCase.On("CreateEndpoint", mock.Anything, reqBody).Return(nil, errors.New("db error"))

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/endpoints", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_LinkEndpointToAccessRight_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights/link", handler.LinkEndpointToAccessRight)

	reqBody := model.LinkEndpointRequest{
		AccessRightID: "1",
		EndpointID:    "1",
	}

	mockUseCase.On("LinkEndpointToAccessRight", mock.Anything, reqBody).Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/access-rights/link", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_LinkEndpointToAccessRight_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights/link", handler.LinkEndpointToAccessRight)

	req, _ := http.NewRequest(http.MethodPost, "/access-rights/link", bytes.NewBufferString(`{"access_right_id":`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUseCase.AssertNotCalled(t, "LinkEndpointToAccessRight", mock.Anything, mock.Anything)
}

func TestAccessHandler_LinkEndpointToAccessRight_ValidationErrors(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights/link", handler.LinkEndpointToAccessRight)

	reqBody := model.LinkEndpointRequest{
		AccessRightID: "",
		EndpointID:    "",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/access-rights/link", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockUseCase.AssertNotCalled(t, "LinkEndpointToAccessRight", mock.Anything, mock.Anything)
}

func TestAccessHandler_LinkEndpointToAccessRight_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights/link", handler.LinkEndpointToAccessRight)

	reqBody := model.LinkEndpointRequest{
		AccessRightID: "1",
		EndpointID:    "1",
	}
	mockUseCase.On("LinkEndpointToAccessRight", mock.Anything, reqBody).Return(errors.New("db error"))

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/access-rights/link", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_DeleteAccessRight_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.DELETE("/access-rights/:id", handler.DeleteAccessRight)

	accessRightID := "1"
	mockUseCase.On("DeleteAccessRight", mock.Anything, accessRightID).Return(nil)

	req, _ := http.NewRequest(http.MethodDelete, "/access-rights/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_DeleteAccessRight_NotFound(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.DELETE("/access-rights/:id", handler.DeleteAccessRight)

	accessRightID := "1"
	mockUseCase.On("DeleteAccessRight", mock.Anything, accessRightID).Return(exception.ErrNotFound)

	req, _ := http.NewRequest(http.MethodDelete, "/access-rights/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_DeleteEndpoint_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.DELETE("/endpoints/:id", handler.DeleteEndpoint)

	endpointID := "1"
	mockUseCase.On("DeleteEndpoint", mock.Anything, endpointID).Return(nil)

	req, _ := http.NewRequest(http.MethodDelete, "/endpoints/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_DeleteEndpoint_NotFound(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.DELETE("/endpoints/:id", handler.DeleteEndpoint)

	endpointID := "1"
	mockUseCase.On("DeleteEndpoint", mock.Anything, endpointID).Return(exception.ErrNotFound)

	req, _ := http.NewRequest(http.MethodDelete, "/endpoints/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_GetEndpointsDynamic_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/endpoints/search", handler.GetEndpointsDynamic)

	filter := querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"Method": {Type: "equals", From: "GET"},
		},
	}
	reqBody, _ := json.Marshal(filter)

	expectedEndpoints := []*model.EndpointResponse{
		{ID: "1", Path: "/test", Method: "GET"},
	}
	mockUseCase.On("GetEndpointsDynamic", mock.Anything, &filter).Return(expectedEndpoints, int64(1), nil)

	req, _ := http.NewRequest(http.MethodPost, "/endpoints/search", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestAccessHandler_GetAccessRightsDynamic_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessController(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights/search", handler.GetAccessRightsDynamic)

	filter := querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"Name": {Type: "contains", From: "Manage"},
		},
	}
	reqBody, _ := json.Marshal(filter)

	expectedResponse := &model.AccessRightListResponse{
		Data: []model.AccessRightResponse{
			{ID: "1", Name: "Manage Users"},
		},
	}
	mockUseCase.On("GetAccessRightsDynamic", mock.Anything, &filter).Return(expectedResponse, int64(1), nil)

	req, _ := http.NewRequest(http.MethodPost, "/access-rights/search", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUseCase.AssertExpectations(t)
}
