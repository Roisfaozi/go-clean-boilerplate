package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.POST("/register", controller.RegisterUser)
	}
}

func RegisterAuthenticatedRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/me", controller.GetCurrentUser)
		userGroup.PUT("/me", controller.UpdateUser)
		userGroup.PATCH("/me/avatar", controller.UpdateAvatar)
	}
}

// RegisterAuthorizedRoutes registers the routes that require rigorous authorization (RBAC).
func RegisterAuthorizedRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("", controller.GetAllUsers)
		userGroup.POST("/search", controller.GetUsersDynamic)
		userGroup.GET("/:id", controller.GetUserByID)
		userGroup.PATCH("/:id/status", controller.UpdateUserStatus)
		userGroup.DELETE("/:id", controller.DeleteUser)
	}
}
