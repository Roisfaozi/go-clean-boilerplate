package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers the public authentication routes.
func RegisterPublicRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/refresh", handler.RefreshToken)
	}
}

// RegisterAuthenticatedRoutes registers auth routes that require authentication.
func RegisterAuthenticatedRoutes(router *gin.RouterGroup, authHandler *AuthHandler) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/logout", authHandler.Logout)
	}
}

