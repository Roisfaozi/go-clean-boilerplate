package http

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/user/delivery/http/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(router *gin.RouterGroup, userHandler *UserHandler) {
	userGroup := router.Group("/users")
	{
		// Public routes
		userGroup.POST("/register", userHandler.RegisterUser)

		// Protected routes (require authentication)
		authGroup := userGroup.Group("")
		authGroup.Use(middleware.AuthMiddleware())
		{
			authGroup.GET("/me", userHandler.GetCurrentUser)
			authGroup.PUT("/me", userHandler.UpdateUser)
		}
	}

	// Auth routes
	authGroup := router.Group("/auth")
	{
		// Protected routes (require authentication)
		authProtected := authGroup.Group("")
		authProtected.Use(middleware.AuthMiddleware())
		{
			authProtected.POST("/logout", userHandler.LogoutUser)
		}
	}
}
