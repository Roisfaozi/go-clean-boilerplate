package test_test

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
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestAccessControllerWithValidator(mockUseCase *mocks.MockIAccessUseCase) *accessHandler.AccessController {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	return accessHandler.NewAccessController(mockUseCase, v, log)
}

func TestAccessHandler_CreateAccessRight_XSS(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessControllerWithValidator(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/access-rights", handler.CreateAccessRight)

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert(1)>",
	}

	for _, payload := range xssPayloads {
		t.Run("XSS Payload: "+payload, func(t *testing.T) {
			reqBody := model.CreateAccessRightRequest{
				Name:        payload,
				Description: "Valid description",
			}

			bodyBytes, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest(http.MethodPost, "/access-rights", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			// Ensure UseCase is NOT called
			mockUseCase.AssertNotCalled(t, "CreateAccessRight", mock.Anything, mock.Anything)
		})
	}
}

func TestAccessHandler_CreateEndpoint_XSS(t *testing.T) {
	mockUseCase := new(mocks.MockIAccessUseCase)
	handler := newTestAccessControllerWithValidator(mockUseCase)
	router := setupAccessTestRouter()
	router.POST("/endpoints", handler.CreateEndpoint)

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
	}

	for _, payload := range xssPayloads {
		t.Run("XSS Payload: "+payload, func(t *testing.T) {
			reqBody := model.CreateEndpointRequest{
				Path:   payload,
				Method: "GET",
			}

			bodyBytes, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest(http.MethodPost, "/endpoints", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
			// Ensure UseCase is NOT called
			mockUseCase.AssertNotCalled(t, "CreateEndpoint", mock.Anything, mock.Anything)
		})
	}
}
