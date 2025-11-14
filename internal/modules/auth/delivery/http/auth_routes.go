package http

import (
	"github.com/Roisfaozi/casbin-db/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers auth routes that do not require authentication.
func RegisterPublicRoutes(router *gin.RouterGroup, authHandler *AuthHandler) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}
}

// RegisterAuthenticatedRoutes registers auth routes that require authentication.
func RegisterAuthenticatedRoutes(router *gin.RouterGroup, authHandler *AuthHandler) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/logout", authHandler.Logout)
	}
}

