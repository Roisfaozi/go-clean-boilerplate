package middleware

import (
	"errors"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/sirupsen/logrus"
)

// HTTPUserStatusMiddleware verifies the user account is currently active.
func HTTPUserStatusMiddleware(userRepo userRepository.UserRepository, log *logrus.Logger) delivery.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := delivery.GetContextString(r.Context(), delivery.UserIDKey)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, errors.New("user context not found"), "unauthorized")
				return
			}

			user, err := userRepo.FindByID(r.Context(), userID)
			if err != nil {
				if log != nil {
					log.WithError(err).Errorf("Failed to fetch user status for ID: %s", userID)
				}
				response.WriteError(w, http.StatusInternalServerError, errors.New("failed to verify user status"), "internal server error")
				return
			}

			switch user.Status {
			case entity.UserStatusSuspended:
				response.WriteError(w, http.StatusForbidden, errors.New(suspendedUserMessage), "account suspended")
				return
			case entity.UserStatusBanned:
				response.WriteError(w, http.StatusForbidden, errors.New(bannedUserMessage), "account banned")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
