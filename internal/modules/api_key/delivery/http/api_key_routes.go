package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterApiKeyRoutes(r *gin.RouterGroup, controller *ApiKeyController, authMiddleware *middleware.AuthMiddleware, tenantMiddleware *middleware.TenantMiddleware, idempotencyMiddleware gin.HandlerFunc) {
	apiKeys := r.Group("/api-keys")
	apiKeys.Use(authMiddleware.ValidateToken())
	apiKeys.Use(tenantMiddleware.RequireOrganization())
	{
		if idempotencyMiddleware != nil {
			apiKeys.POST("", idempotencyMiddleware, delivery.AdaptHTTPHandler(controller.HTTPCreate))
		} else {
			apiKeys.POST("", delivery.AdaptHTTPHandler(controller.HTTPCreate))
		}
		apiKeys.GET("", delivery.AdaptHTTPHandler(controller.HTTPList))
		apiKeys.DELETE("/:id", delivery.AdaptHTTPHandler(controller.HTTPRevoke))
	}
}
