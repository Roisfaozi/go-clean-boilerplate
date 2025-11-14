package http

import (
	"github.com/Roisfaozi/casbin-db/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(router *gin.RouterGroup, userHandler *UserHandler, authMiddleware *middleware.AuthMiddleware) {
	userGroup := router.Group("/users")
	{
		userGroup.POST("/register", userHandler.RegisterUser)

		authGroup := userGroup.Group("")
		authGroup.Use(authMiddleware.ValidateToken())
		{
			authGroup.GET("/me", userHandler.GetCurrentUser)
			authGroup.PUT("/me", userHandler.UpdateUser)
		}
	}

	// Auth routes
	authGroup := router.Group("/auth")
	{
		authProtected := authGroup.Group("")
		authProtected.Use(authMiddleware.ValidateToken())
		{
			authProtected.POST("/logout", userHandler.LogoutUser)
		}
	}
}
