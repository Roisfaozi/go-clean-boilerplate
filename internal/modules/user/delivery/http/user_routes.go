package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers the routes that do not require authorization.
//
// router: the router group to register the routes on.
// userHandler: the handler for the user routes.
func RegisterPublicRoutes(router *gin.RouterGroup, userHandler *UserHandler) {
	userGroup := router.Group("/users")
	{
		// register the route for user registration
		//
		// @Summary      Register a new user
		// @Description  Creates a new user account.
		// @Tags         users
		// @Accept       json
		// @Produce      json
		// @Param        request body model.RegisterUserRequest true "User Registration Details"
		userGroup.POST("/register", userHandler.RegisterUser)
	}
}

// RegisterAuthorizedRoutes registers the routes that require authorization.
//
// router: the router group to register the routes on.
// userHandler: the handler for the user routes.
func RegisterAuthorizedRoutes(router *gin.RouterGroup, userHandler *UserHandler) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/me", userHandler.GetCurrentUser)
		userGroup.PUT("/me", userHandler.UpdateUser)
		userGroup.GET("/", userHandler.GetAllUsers)
		userGroup.POST("/search", userHandler.GetUsersDynamic)
		userGroup.GET("/:id", userHandler.GetUserByID)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
	}
}
