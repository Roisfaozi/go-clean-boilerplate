package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apiKeyModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	apiKeyMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPIKeyMiddleware_Authenticate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logrus.New()

	mockUseCase := new(apiKeyMocks.MockApiKeyUseCase)
	mockUserRepo := new(userMocks.MockUserRepository)

	mw := NewAPIKeyMiddleware(mockUseCase, mockUserRepo, log)

	t.Run("Valid API Key", func(t *testing.T) {
		r := gin.New()
		r.Use(mw.Authenticate())
		r.GET("/test", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			c.String(http.StatusOK, userID.(string))
		})

		key := "sk_live_valid_key"
		keyIdentity := &apiKeyModel.ApiKeyIdentity{
			UserID:         "user-123",
			OrganizationID: "org-456",
			Username:       "api_user",
		}

		mockUseCase.On("Authenticate", mock.Anything, key).Return(keyIdentity, nil)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", key)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "user-123", w.Body.String())
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		r := gin.New()
		r.Use(mw.Authenticate())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "should not reach here")
		})

		key := "sk_live_invalid"
		mockUseCase.On("Authenticate", mock.Anything, key).Return(nil, assert.AnError)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", key)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
