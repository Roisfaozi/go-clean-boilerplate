package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	userHandler "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestUserXSSValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.FatalLevel)
	tests := []struct {
		name         string
		method       string
		url          string
		payload      interface{}
		expectedCode int
	}{
		{
			name:   "RegisterUser XSS in Name",
			method: "POST",
			url:    "/users",
			payload: model.RegisterUserRequest{
				Username: "testuser",
				Password: "password123",
				Name:     "<script>alert(1)</script>",
				Email:    "test@example.com",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "RegisterUser XSS in Username",
			method: "POST",
			url:    "/users",
			payload: model.RegisterUserRequest{
				Username: "<img src=x onerror=alert(1)>",
				Password: "password123",
				Name:     "Test User",
				Email:    "test@example.com",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:   "UpdateUser XSS in Name",
			method: "PUT",
			url:    "/users/1",
			payload: model.UpdateUserRequest{
				Name:     "<iframe src='javascript:alert(1)'></iframe>",
				Username: "testuser",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := new(mocks.MockUserUseCase)
			controller := userHandler.NewUserController(mockUC, logger, v)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest(tt.method, tt.url, bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Set("user_id", "1")

			if tt.method == "POST" {
				controller.RegisterUser(c)
			} else {
				controller.UpdateUser(c)
			}

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
