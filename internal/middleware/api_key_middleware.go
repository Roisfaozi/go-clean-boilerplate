package middleware

import (
	"errors"

	apiKeyUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type APIKeyMiddleware struct {
	ApiKeyUseCase apiKeyUsecase.ApiKeyUseCase
	UserRepo      userRepository.UserRepository
	Log           *logrus.Logger
}

func NewAPIKeyMiddleware(apiKeyUseCase apiKeyUsecase.ApiKeyUseCase, userRepo userRepository.UserRepository, log *logrus.Logger) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		ApiKeyUseCase: apiKeyUseCase,
		UserRepo:      userRepo,
		Log:           log,
	}
}

func (m *APIKeyMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.Next() // Allow other auth methods (JWT) to handle it
			return
		}

		keyEntity, err := m.ApiKeyUseCase.Authenticate(c.Request.Context(), apiKey)
		if err != nil {
			m.Log.WithError(err).Warn("API Key authentication failed")
			response.Unauthorized(c, err, "unauthorized")
			c.Abort()
			return
		}

		// Get full user details
		user, err := m.UserRepo.FindByID(c.Request.Context(), keyEntity.UserID)
		if err != nil {
			m.Log.WithError(err).Error("Failed to fetch user associated with API Key")
			response.Unauthorized(c, errors.New("invalid api key owner"), "unauthorized")
			c.Abort()
			return
		}

		// Inject into context
		c.Set("user_id", user.ID)
		c.Set("organization_id", keyEntity.OrganizationID)
		c.Set("username", user.Username)
		c.Set("auth_method", "api_key")

		c.Next()
	}
}
