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
//   - controller: the *AuthController to handle requests
func RegisterPublicRoutes(router *gin.RouterGroup, controller *AuthController) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", controller.Login)
		authGroup.POST("/refresh", controller.RefreshToken)
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
//   - controller: the *AuthController to handle requests
func RegisterAuthenticatedRoutes(router *gin.RouterGroup, controller *AuthController) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/logout", controller.Logout)
	}
}