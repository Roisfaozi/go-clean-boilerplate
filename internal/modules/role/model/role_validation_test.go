package model_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	roleHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoleXSSValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	logger := logrus.New()

	tests := []struct {
		name         string
		method       string
		url          string
		payload      interface{}
		expectedCode int
		setupMock    func(*mocks.MockRoleUseCase)
	}{
		{
			name:   "CreateRole XSS in Name",
			method: "POST",
			url:    "/roles",
			payload: model.CreateRoleRequest{
				Name:        "<script>alert(1)</script>",
				Description: "A role",
			},
			expectedCode: http.StatusCreated,
			setupMock: func(m *mocks.MockRoleUseCase) {
				// Sanitized: <script>alert(1)</script> -> &lt;script&gt;alert(1)&lt;/script&gt;
				m.On("Create", mock.Anything, &model.CreateRoleRequest{Name: "&lt;script&gt;alert(1)&lt;/script&gt;", Description: "A role"}).Return(&model.RoleResponse{ID: "1", Name: "&lt;script&gt;alert(1)&lt;/script&gt;"}, nil)
			},
		},
		{
			name:   "CreateRole XSS in Description",
			method: "POST",
			url:    "/roles",
			payload: model.CreateRoleRequest{
				Name:        "admin",
				Description: "<img src=x onerror=alert(2)>",
			},
			expectedCode: http.StatusCreated,
			setupMock: func(m *mocks.MockRoleUseCase) {
				// Sanitized: <img src=x onerror=alert(2)> -> &lt;img src=x onerror=alert(2)&gt;
				m.On("Create", mock.Anything, &model.CreateRoleRequest{Name: "admin", Description: "&lt;img src=x onerror=alert(2)&gt;"}).Return(&model.RoleResponse{ID: "1", Name: "admin"}, nil)
			},
		},
		{
			name:   "UpdateRole XSS in Description",
			method: "PUT",
			url:    "/roles/1",
			payload: model.UpdateRoleRequest{
				Description: "<iframe src='javascript:alert(3)'></iframe>",
			},
			expectedCode: http.StatusOK,
			setupMock: func(m *mocks.MockRoleUseCase) {
				m.On("Update", mock.Anything, "1", &model.UpdateRoleRequest{Description: "&lt;iframe src=&#39;javascript:alert(3)&#39;&gt;&lt;/iframe&gt;"}).Return(&model.RoleResponse{ID: "1", Description: "&lt;iframe src=&#39;javascript:alert(3)&#39;&gt;&lt;/iframe&gt;"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mocks.MockRoleUseCase)
			if tt.setupMock != nil {
				tt.setupMock(mockUC)
			}
			controller := roleHandler.NewRoleController(mockUC, logger, v)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest(tt.method, tt.url, bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			if tt.method == "PUT" {
				c.Params = []gin.Param{{Key: "id", Value: "1"}}
			}

			if tt.method == "POST" {
				controller.Create(c)
			} else {
				controller.Update(c)
			}

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
