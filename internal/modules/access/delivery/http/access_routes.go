package http

import "github.com/gin-gonic/gin"

// RegisterAccessRoutes registers the access-related HTTP routes.
//
// RegisterAccessRoutes sets up the routes for creating access rights and
// endpoints. It takes a *gin.RouterGroup as the first argument and an
// *AccessController as the second argument. The *gin.RouterGroup is used to add
// routes to a specific group of routes, and the *AccessController is used to
// handle the requests to those routes.
//
// The routes registered by this function are:
//   - POST /access-rights: creates a new access right
//   - GET /access-rights: retrieves a list of all available access rights
//   - POST /access-rights/link: links an endpoint to an access right
//   - POST /endpoints: creates a new API endpoint
//
// Parameters:
//   - router: the *gin.RouterGroup to add routes to
//   - handler: the *AccessController to handle requests
func RegisterAccessRoutes(router *gin.RouterGroup, controller *AccessController) {
	accessGroup := router.Group("/access-rights")
	{
		accessGroup.POST("", controller.CreateAccessRight)
		accessGroup.GET("", controller.GetAllAccessRights)
		accessGroup.POST("/search", controller.GetAccessRightsDynamic)
		accessGroup.DELETE("/:id", controller.DeleteAccessRight)
		accessGroup.POST("/link", controller.LinkEndpointToAccessRight)
		accessGroup.POST("/unlink", controller.UnlinkEndpointFromAccessRight)
	}

	endpointGroup := router.Group("/endpoints")
	{
		endpointGroup.POST("", controller.CreateEndpoint)
		endpointGroup.POST("/search", controller.GetEndpointsDynamic)
		endpointGroup.DELETE("/:id", controller.DeleteEndpoint)
	}
}
