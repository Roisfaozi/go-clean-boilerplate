package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	accessHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAccessRight_XSS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		payload      model.CreateAccessRightRequest
		expectedCode int
	}{
		{
			name: "XSS in Name",
			payload: model.CreateAccessRightRequest{
				Name:        "<script>alert(1)</script>",
				Description: "Valid Description",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "XSS in Description",
			payload: model.CreateAccessRightRequest{
				Name:        "Valid Name",
				Description: "<img src=x onerror=alert(1)>",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Safe content",
			payload: model.CreateAccessRightRequest{
				Name:        "Safe Name",
				Description: "Safe Description",
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccessUseCase := mocks.NewMockIAccessUseCase(t)

			v := validator.New()
			_ = validation.RegisterCustomValidations(v)
			logger := logrus.New()

			controller := accessHandler.NewAccessController(mockAccessUseCase, v, logger)

			if tt.expectedCode == http.StatusCreated {
				mockAccessUseCase.On("CreateAccessRight", mock.Anything, mock.Anything).Return(nil, nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest("POST", "/access-rights", bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.CreateAccessRight(c)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestCreateEndpoint_XSS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		payload      model.CreateEndpointRequest
		expectedCode int
	}{
		{
			name: "XSS in Path",
			payload: model.CreateEndpointRequest{
				Path:   "/api/<script>alert(1)</script>",
				Method: "GET",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "XSS in Method",
			payload: model.CreateEndpointRequest{
				Path:   "/api/valid",
				Method: "<script>",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Safe content",
			payload: model.CreateEndpointRequest{
				Path:   "/api/valid",
				Method: "POST",
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccessUseCase := mocks.NewMockIAccessUseCase(t)

			v := validator.New()
			_ = validation.RegisterCustomValidations(v)
			logger := logrus.New()

			controller := accessHandler.NewAccessController(mockAccessUseCase, v, logger)

			if tt.expectedCode == http.StatusCreated {
				mockAccessUseCase.On("CreateEndpoint", mock.Anything, mock.Anything).Return(nil, nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest("POST", "/endpoints", bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.CreateEndpoint(c)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
