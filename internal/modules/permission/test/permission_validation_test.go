package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	permissionHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssignRole_XSS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		payload      model.AssignRoleRequest
		expectedCode int
	}{
		{
			name: "XSS in UserID",
			payload: model.AssignRoleRequest{
				UserID: "<script>alert(1)</script>",
				Role:   "admin",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "XSS in Role",
			payload: model.AssignRoleRequest{
				UserID: "user1",
				Role:   "<img src=x onerror=alert(1)>",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Safe content",
			payload: model.AssignRoleRequest{
				UserID: "user1",
				Role:   "admin",
				Domain: "global",
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockIPermissionUseCase(t)

			v := validator.New()
			_ = validation.RegisterCustomValidations(v)
			logger := logrus.New()

			controller := permissionHandler.NewPermissionController(mockUseCase, logger, v)

			if tt.expectedCode == http.StatusOK {
				mockUseCase.On("AssignRoleToUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest("POST", "/roles/assign", bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.AssignRole(c)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestGrantPermission_XSS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		payload      model.GrantPermissionRequest
		expectedCode int
	}{
		{
			name: "XSS in Role",
			payload: model.GrantPermissionRequest{
				Role:   "<script>alert(1)</script>",
				Path:   "/api/resource",
				Method: "GET",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "XSS in Path",
			payload: model.GrantPermissionRequest{
				Role:   "admin",
				Path:   "/api/<script>alert(1)</script>",
				Method: "GET",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "XSS in Method",
			payload: model.GrantPermissionRequest{
				Role:   "admin",
				Path:   "/api/resource",
				Method: "<img src=x>",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Safe content",
			payload: model.GrantPermissionRequest{
				Role:   "admin",
				Path:   "/api/resource",
				Method: "GET",
				Domain: "global",
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockIPermissionUseCase(t)

			v := validator.New()
			_ = validation.RegisterCustomValidations(v)
			logger := logrus.New()

			controller := permissionHandler.NewPermissionController(mockUseCase, logger, v)

			if tt.expectedCode == http.StatusCreated {
				mockUseCase.On("GrantPermissionToRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest("POST", "/permissions/grant", bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.GrantPermission(c)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestUpdatePermission_XSS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		payload      model.UpdatePermissionRequest
		expectedCode int
	}{
		{
			name: "XSS in OldPermission",
			payload: model.UpdatePermissionRequest{
				OldPermission: []string{"role", "/path", "<script>alert(1)</script>"},
				NewPermission: []string{"role", "/path", "GET"},
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "XSS in NewPermission",
			payload: model.UpdatePermissionRequest{
				OldPermission: []string{"role", "/path", "GET"},
				NewPermission: []string{"role", "/api/<img src=x>", "GET"},
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Safe content",
			payload: model.UpdatePermissionRequest{
				OldPermission: []string{"role", "global", "/path", "GET"},
				NewPermission: []string{"role", "global", "/new-path", "POST"},
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := mocks.NewMockIPermissionUseCase(t)

			v := validator.New()
			_ = validation.RegisterCustomValidations(v)
			logger := logrus.New()

			controller := permissionHandler.NewPermissionController(mockUseCase, logger, v)

			if tt.expectedCode == http.StatusOK {
				mockUseCase.On("UpdatePermission", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest("PUT", "/permissions", bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.UpdatePermission(c)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
