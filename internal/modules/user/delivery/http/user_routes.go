package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers the routes that do not require authorization.
//
// router: the router group to register the routes on.
// controller: the controller for the user routes.
func RegisterPublicRoutes(router *gin.RouterGroup, controller *UserController) {
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
		userGroup.POST("/register", controller.RegisterUser)
	}
}

// RegisterAuthorizedRoutes registers the routes that require authorization.
//
// router: the router group to register the routes on.
// controller: the controller for the user routes.
func RegisterAuthorizedRoutes(router *gin.RouterGroup, controller *UserController) {
	userGroup := router.Group("/users")
	{
		userGroup.GET("/me", controller.GetCurrentUser)
		userGroup.PUT("/me", controller.UpdateUser)
		userGroup.GET("/", controller.GetAllUsers)
		userGroup.POST("/search", controller.GetUsersDynamic)
		userGroup.GET("/:id", controller.GetUserByID)
		userGroup.DELETE("/:id", controller.DeleteUser)
	}
}