package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterApiKeyRoutes(r *gin.RouterGroup, controller *ApiKeyController, authMiddleware *middleware.AuthMiddleware, tenantMiddleware *middleware.TenantMiddleware) {
	apiKeys := r.Group("/api-keys")
	apiKeys.Use(authMiddleware.ValidateToken())
	apiKeys.Use(tenantMiddleware.RequireOrganization())
	{
		apiKeys.POST("", controller.Create)
		apiKeys.GET("", controller.List)
		apiKeys.DELETE("/:id", controller.Revoke)
	}
}
