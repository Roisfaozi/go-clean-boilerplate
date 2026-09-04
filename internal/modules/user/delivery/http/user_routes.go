package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.POST("/register", delivery.AdaptHTTPHandler(controller.HTTPRegisterUser))
	}
}

func RegisterAuthenticatedRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/me", delivery.AdaptHTTPHandler(controller.HTTPGetCurrentUser))
		userGroup.PUT("/me", delivery.AdaptHTTPHandler(controller.HTTPUpdateUser))
		userGroup.PATCH("/me/avatar", delivery.AdaptHTTPHandler(controller.HTTPUpdateAvatar))
	}
}

// RegisterAuthorizedRoutes registers the routes that require rigorous authorization (RBAC).
func RegisterAuthorizedRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("", delivery.AdaptHTTPHandler(controller.HTTPGetAllUsers))
		userGroup.POST("/search", delivery.AdaptHTTPHandler(controller.HTTPGetUsersDynamic))
		userGroup.GET("/:id", delivery.AdaptHTTPHandler(controller.HTTPGetUserByID))
		userGroup.PATCH("/:id/status", delivery.AdaptHTTPHandler(controller.HTTPUpdateUserStatus))
		userGroup.DELETE("/:id", delivery.AdaptHTTPHandler(controller.HTTPDeleteUser))
	}
}
