package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers user routes that do not require authentication.
func RegisterPublicRoutes(router *gin.RouterGroup, userHandler *UserHandler) {
	userGroup := router.Group("/users")
	{
		userGroup.POST("/register", userHandler.RegisterUser)
	}
}

// RegisterAuthorizedRoutes registers user routes that require both authentication and authorization.
func RegisterAuthorizedRoutes(router *gin.RouterGroup, userHandler *UserHandler) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/me", userHandler.GetCurrentUser)
		userGroup.PUT("/me", userHandler.UpdateUser)
		userGroup.GET("/", userHandler.GetAllUsers)
		userGroup.GET("/:id", userHandler.GetUserByID)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
	}
}

