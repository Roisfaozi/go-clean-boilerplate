package test

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
		},
		{
			name:   "UpdateRole XSS in Description",
			method: "PUT",
			url:    "/roles/1",
			payload: model.UpdateRoleRequest{
				Description: "<iframe src='javascript:alert(3)'></iframe>",
			},
			expectedCode: http.StatusUnprocessableEntity, // Still 422 because required field becomes empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mocks.MockRoleUseCase)
			controller := roleHandler.NewRoleController(mockUC, logger, v)

			// Setup expectations for success cases
			if tt.expectedCode == http.StatusCreated {
				mockUC.On("Create", mock.Anything, mock.Anything).Return(&model.RoleResponse{ID: "uuid"}, nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest(tt.method, tt.url, bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.method == "POST" {
				controller.Create(c)
			} else {
				controller.Update(c)
			}

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
