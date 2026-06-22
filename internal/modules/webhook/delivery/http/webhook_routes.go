package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterWebhookRoutes(r *gin.RouterGroup, controller *WebhookController, apiKeyMiddleware *middleware.APIKeyMiddleware, idempotencyMiddleware gin.HandlerFunc) {
	webhooks := r.Group("/webhooks")
	{
		webhooks.Use(apiKeyMiddleware.RequireScopes("webhook:manage"))
		if idempotencyMiddleware != nil {
			webhooks.POST("", idempotencyMiddleware, controller.Create)
		} else {
			webhooks.POST("", controller.Create)
		}
		webhooks.GET("", controller.FindByOrganization)
		webhooks.GET("/:id", controller.FindByID)
		webhooks.PUT("/:id", controller.Update)
		webhooks.DELETE("/:id", controller.Delete)
		webhooks.GET("/:id/logs", controller.GetLogs)
	}
}
