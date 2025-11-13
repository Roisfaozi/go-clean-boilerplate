package http

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers all auth-related routes
func RegisterAuthRoutes(router *gin.RouterGroup, authHandler *AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	authGroup := router.Group("/auth")
	{
		// Public routes
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)

		// Protected routes
		protected := authGroup.Group("")
		protected.Use(authMiddleware.ValidateToken())
		{
			protected.POST("/logout", authHandler.Logout)
		}
	}
}
