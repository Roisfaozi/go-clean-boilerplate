package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	accessHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/querybuilder"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type accessControllerDeps struct {
	UseCase *mocks.MockIAccessUseCase
}

func setupAccessControllerTest() (*accessControllerDeps, *gin.Engine, *accessHttp.AccessController) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	deps := &accessControllerDeps{
		UseCase: new(mocks.MockIAccessUseCase),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)

	validator := config.NewValidator()
	controller := accessHttp.NewAccessController(deps.UseCase, validator, log)

	return deps, engine, controller
}

func TestAccessController_CreateAccessRight(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.POST("/access-rights", controller.CreateAccessRight)

		reqBody := model.CreateAccessRightRequest{Name: "Admin", Description: "Full access"}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/access-rights", bytes.NewBuffer(jsonValue))

		mockRes := &model.AccessRightResponse{ID: "ar-1", Name: "Admin"}
		deps.UseCase.EXPECT().CreateAccessRight(mock.Anything, mock.AnythingOfType("model.CreateAccessRightRequest")).Return(mockRes, nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, engine, controller := setupAccessControllerTest()
		engine.POST("/access-rights", controller.CreateAccessRight)

		reqBody := model.CreateAccessRightRequest{} // Missing Name
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/access-rights", bytes.NewBuffer(jsonValue))

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestAccessController_CreateEndpoint(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.POST("/endpoints", controller.CreateEndpoint)

		reqBody := model.CreateEndpointRequest{Path: "/api/test", Method: "GET"}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/endpoints", bytes.NewBuffer(jsonValue))

		mockRes := &model.EndpointResponse{ID: "ep-1", Path: "/api/test", Method: "GET"}
		deps.UseCase.EXPECT().CreateEndpoint(mock.Anything, mock.AnythingOfType("model.CreateEndpointRequest")).Return(mockRes, nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("Validation Error", func(t *testing.T) {
		_, engine, controller := setupAccessControllerTest()
		engine.POST("/endpoints", controller.CreateEndpoint)

		reqBody := model.CreateEndpointRequest{} // Missing required fields
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/endpoints", bytes.NewBuffer(jsonValue))

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestAccessController_LinkUnlinkEndpoint(t *testing.T) {
	t.Run("Link Success", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.POST("/access-rights/link", controller.LinkEndpointToAccessRight)

		reqBody := model.LinkEndpointRequest{AccessRightID: "ar-1", EndpointID: "ep-1"}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/access-rights/link", bytes.NewBuffer(jsonValue))

		deps.UseCase.EXPECT().LinkEndpointToAccessRight(mock.Anything, mock.AnythingOfType("model.LinkEndpointRequest")).Return(nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("Unlink Success", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.POST("/access-rights/unlink", controller.UnlinkEndpointFromAccessRight)

		reqBody := model.LinkEndpointRequest{AccessRightID: "ar-1", EndpointID: "ep-1"}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/access-rights/unlink", bytes.NewBuffer(jsonValue))

		deps.UseCase.EXPECT().UnlinkEndpointFromAccessRight(mock.Anything, mock.AnythingOfType("model.LinkEndpointRequest")).Return(nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		deps.UseCase.AssertExpectations(t)
	})
}

func TestAccessController_DeleteAccessRightAndEndpoint(t *testing.T) {
	t.Run("DeleteAccessRight Success", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.DELETE("/access-rights/:id", controller.DeleteAccessRight)

		req, _ := http.NewRequest(http.MethodDelete, "/access-rights/ar-1", nil)
		deps.UseCase.EXPECT().DeleteAccessRight(mock.Anything, "ar-1").Return(nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("DeleteAccessRight NotFound", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.DELETE("/access-rights/:id", controller.DeleteAccessRight)

		req, _ := http.NewRequest(http.MethodDelete, "/access-rights/ar-1", nil)
		deps.UseCase.EXPECT().DeleteAccessRight(mock.Anything, "ar-1").Return(exception.ErrNotFound).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("DeleteEndpoint Success", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.DELETE("/endpoints/:id", controller.DeleteEndpoint)

		req, _ := http.NewRequest(http.MethodDelete, "/endpoints/ep-1", nil)
		deps.UseCase.EXPECT().DeleteEndpoint(mock.Anything, "ep-1").Return(nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		deps.UseCase.AssertExpectations(t)
	})
}

func TestAccessController_DynamicEndpoints(t *testing.T) {
	t.Run("GetEndpointsDynamic Success", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.POST("/endpoints/search", controller.GetEndpointsDynamic)

		reqBody := querybuilder.DynamicFilter{Page: 1, PageSize: 10}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/endpoints/search", bytes.NewBuffer(jsonValue))

		mockRes := []*model.EndpointResponse{{ID: "ep-1"}}
		deps.UseCase.EXPECT().GetEndpointsDynamic(mock.Anything, mock.AnythingOfType("*querybuilder.DynamicFilter")).Return(mockRes, int64(1), nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("GetAccessRightsDynamic Success", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.POST("/access-rights/search", controller.GetAccessRightsDynamic)

		reqBody := querybuilder.DynamicFilter{Page: 1, PageSize: 10}
		jsonValue, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest(http.MethodPost, "/access-rights/search", bytes.NewBuffer(jsonValue))

		mockRes := &model.AccessRightListResponse{Data: []model.AccessRightResponse{{ID: "ar-1"}}}
		deps.UseCase.EXPECT().GetAccessRightsDynamic(mock.Anything, mock.AnythingOfType("*querybuilder.DynamicFilter")).Return(mockRes, int64(1), nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		deps.UseCase.AssertExpectations(t)
	})
}

func TestAccessController_GetAllAccessRights(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.GET("/access-rights", controller.GetAllAccessRights)

		req, _ := http.NewRequest(http.MethodGet, "/access-rights", nil)

		mockRes := &model.AccessRightListResponse{Data: []model.AccessRightResponse{{ID: "ar-1"}}}
		deps.UseCase.EXPECT().GetAllAccessRights(mock.Anything).Return(mockRes, nil).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		deps.UseCase.AssertExpectations(t)
	})

	t.Run("UseCase Error", func(t *testing.T) {
		deps, engine, controller := setupAccessControllerTest()
		engine.GET("/access-rights", controller.GetAllAccessRights)

		req, _ := http.NewRequest(http.MethodGet, "/access-rights", nil)

		deps.UseCase.EXPECT().GetAllAccessRights(mock.Anything).Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		deps.UseCase.AssertExpectations(t)
	})
}