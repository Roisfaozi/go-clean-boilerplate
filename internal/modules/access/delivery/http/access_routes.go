package http

import "github.com/gin-gonic/gin"

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
