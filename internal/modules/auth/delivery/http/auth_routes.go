package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers the public routes for authentication.
func RegisterPublicRoutes(router *gin.RouterGroup, controller *AuthController) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", controller.Login)
		authGroup.POST("/refresh", controller.RefreshToken)
		authGroup.POST("/forgot-password", controller.ForgotPassword)
		authGroup.POST("/reset-password", controller.ResetPassword)
	}
}

// RegisterAuthenticatedRoutes registers the routes for authenticated users.
func RegisterAuthenticatedRoutes(router *gin.RouterGroup, controller *AuthController) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/logout", controller.Logout)
	}
}
