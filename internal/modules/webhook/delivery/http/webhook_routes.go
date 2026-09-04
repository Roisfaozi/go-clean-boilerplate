package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterWebhookRoutes(r *gin.RouterGroup, controller *WebhookController, apiKeyMiddleware *middleware.APIKeyMiddleware, idempotencyMiddleware gin.HandlerFunc) {
	webhooks := r.Group("/webhooks")
	{
		webhooks.Use(apiKeyMiddleware.RequireScopes("webhook:manage"))
		if idempotencyMiddleware != nil {
			webhooks.POST("", idempotencyMiddleware, delivery.AdaptHTTPHandler(controller.HTTPCreate))
		} else {
			webhooks.POST("", delivery.AdaptHTTPHandler(controller.HTTPCreate))
		}
		webhooks.GET("", delivery.AdaptHTTPHandler(controller.HTTPFindByOrganization))
		webhooks.GET("/:id", delivery.AdaptHTTPHandler(controller.HTTPFindByID))
		webhooks.PUT("/:id", delivery.AdaptHTTPHandler(controller.HTTPUpdate))
		webhooks.DELETE("/:id", delivery.AdaptHTTPHandler(controller.HTTPDelete))
		webhooks.GET("/:id/logs", delivery.AdaptHTTPHandler(controller.HTTPGetLogs))
	}
}
