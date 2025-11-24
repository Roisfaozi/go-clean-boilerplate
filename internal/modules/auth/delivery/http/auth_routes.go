package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers the public routes for authentication.
//
// RegisterPublicRoutes adds the following routes to the provided
// *gin.RouterGroup:
//   - POST /auth/login: creates a new access token
//   - POST /auth/refresh: refreshes an existing access token
//
// Parameters:
//   - router: the *gin.RouterGroup to add routes to
//   - handler: the *AuthHandler to handle requests
func RegisterPublicRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", handler.Login)
		authGroup.POST("/refresh", handler.RefreshToken)
	}
}

// RegisterAuthenticatedRoutes registers the routes for authenticated users.
//
// RegisterAuthenticatedRoutes adds the following routes to the provided
// *gin.RouterGroup:
//   - POST /auth/logout: logs out the current user
//
// Parameters:
//   - router: the *gin.RouterGroup to add routes to
//   - authHandler: the *AuthHandler to handle requests
func RegisterAuthenticatedRoutes(router *gin.RouterGroup, authHandler *AuthHandler) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/logout", authHandler.Logout)
	}
}
