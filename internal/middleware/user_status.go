package middleware

import (
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// UserStatusMiddleware checks if the user account is active.
//
// It assumes that the user_id has already been set in the context by AuthMiddleware.
// If the user_id is not found in the context or the user status is inactive,
// it will return a 403 Forbidden response.
// If there is an error fetching the user status, it will return a 500 Internal Server Error response.
// Otherwise, it will call the next handler in the chain.
func UserStatusMiddleware(userRepo userRepository.UserRepository, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, errors.New("user context not found"), "unauthorized")
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(c.Request.Context(), userID.(string))
		if err != nil {
			log.WithError(err).Errorf("Failed to fetch user status for ID: %s", userID)
			response.InternalServerError(c, errors.New("failed to verify user status"), "internal server error")
			c.Abort()
			return
		}

		if user.Status != entity.UserStatusActive {
			log.Warnf("Access denied for %s user: %s", user.Status, userID)

			msg := "Your account has been banned. Please contact support."
			if user.Status == entity.UserStatusSuspended {
				msg = "Your account has been suspended temporarily. Please contact support."
			}

			response.Forbidden(c, errors.New("forbidden"), msg)
			c.Abort()
			return
		}

		c.Next()
	}
}
