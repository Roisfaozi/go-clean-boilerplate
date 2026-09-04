package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/gin-gonic/gin"
)

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
func RegisterAccessRoutes(router *gin.RouterGroup, controller *AccessController, idempotencyMiddleware gin.HandlerFunc) {
	accessGroup := router.Group("/access-rights")
	{
		if idempotencyMiddleware != nil {
			accessGroup.POST("", idempotencyMiddleware, delivery.AdaptHTTPHandler(controller.HTTPCreateAccessRight))
		} else {
			accessGroup.POST("", delivery.AdaptHTTPHandler(controller.HTTPCreateAccessRight))
		}
		accessGroup.GET("", delivery.AdaptHTTPHandler(controller.HTTPGetAllAccessRights))
		accessGroup.POST("/search", delivery.AdaptHTTPHandler(controller.HTTPGetAccessRightsDynamic))
		accessGroup.DELETE("/:id", delivery.AdaptHTTPHandler(controller.HTTPDeleteAccessRight))
		accessGroup.POST("/link", delivery.AdaptHTTPHandler(controller.HTTPLinkEndpointToAccessRight))
		accessGroup.POST("/unlink", delivery.AdaptHTTPHandler(controller.HTTPUnlinkEndpointFromAccessRight))
	}

	endpointGroup := router.Group("/endpoints")
	{
		if idempotencyMiddleware != nil {
			endpointGroup.POST("", idempotencyMiddleware, delivery.AdaptHTTPHandler(controller.HTTPCreateEndpoint))
		} else {
			endpointGroup.POST("", delivery.AdaptHTTPHandler(controller.HTTPCreateEndpoint))
		}
		endpointGroup.POST("/search", delivery.AdaptHTTPHandler(controller.HTTPGetEndpointsDynamic))
		endpointGroup.DELETE("/:id", delivery.AdaptHTTPHandler(controller.HTTPDeleteEndpoint))
	}
}
