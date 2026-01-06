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

// RegisterAuthorizedRoutes registers the routes that require authorization.
func RegisterAuthorizedRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/me", controller.GetCurrentUser)
		userGroup.PUT("/me", controller.UpdateUser)
		userGroup.GET("/", controller.GetAllUsers)
		userGroup.POST("/search", controller.GetUsersDynamic)
		userGroup.GET("/:id", controller.GetUserByID)
		userGroup.PATCH("/:id/status", controller.UpdateUserStatus) // Endpoint baru
		userGroup.DELETE("/:id", controller.DeleteUser)
	}
}
