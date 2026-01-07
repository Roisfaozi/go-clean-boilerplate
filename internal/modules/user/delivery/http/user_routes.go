package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers the routes that do not require authorization.
func RegisterPublicRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.POST("/register", controller.RegisterUser)
	}
}

// RegisterAuthenticatedRoutes registers routes that require the user to be logged in.
// These routes do NOT require RBAC or Active status checks (e.g. they might be accessible even if status is 'suspended' depending on policy,
// but definitely don't require Admin privileges).
func RegisterAuthenticatedRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/me", controller.GetCurrentUser)
		userGroup.PUT("/me", controller.UpdateUser)
	}
}

// RegisterAuthorizedRoutes registers the routes that require authorization (RBAC + Status Check).
// These are sensitive administrative endpoints.
func RegisterAuthorizedRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/", controller.GetAllUsers)
		userGroup.POST("/search", controller.GetUsersDynamic)
		userGroup.GET("/:id", controller.GetUserByID)
		userGroup.PATCH("/:id/status", controller.UpdateUserStatus)
		userGroup.DELETE("/:id", controller.DeleteUser)
	}
}
