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
// It assumes the user_id has already been set in the context by AuthMiddleware.
func UserStatusMiddleware(userRepo userRepository.UserRepository, log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			// This should not happen if AuthMiddleware is called first
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
