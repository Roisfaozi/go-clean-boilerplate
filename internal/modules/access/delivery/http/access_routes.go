package http

import "github.com/gin-gonic/gin"

// RegisterAccessRoutes registers the access-related HTTP routes.
//
// RegisterAccessRoutes sets up the routes for creating access rights and
// endpoints. It takes a *gin.RouterGroup as the first argument and an
// *AccessHandler as the second argument. The *gin.RouterGroup is used to add
// routes to a specific group of routes, and the *AccessHandler is used to
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
//   - handler: the *AccessHandler to handle requests
func RegisterAccessRoutes(router *gin.RouterGroup, handler *AccessHandler) {
	accessGroup := router.Group("/access-rights")
	{
		accessGroup.POST("", handler.CreateAccessRight)
		accessGroup.GET("", handler.GetAllAccessRights)
		accessGroup.POST("/link", handler.LinkEndpointToAccessRight)
	}

	endpointGroup := router.Group("/endpoints")
	{
		endpointGroup.POST("", handler.CreateEndpoint)
	}
}
