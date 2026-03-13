package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterWebhookRoutes(r *gin.RouterGroup, controller *WebhookController, authMiddleware gin.HandlerFunc, casbinMiddleware gin.HandlerFunc) {
	webhooks := r.Group("/webhooks")
	webhooks.Use(authMiddleware, casbinMiddleware)
	{
		webhooks.POST("", controller.Create)
		webhooks.GET("", controller.FindByOrganization)
		webhooks.GET("/:id", controller.FindByID)
		webhooks.PUT("/:id", controller.Update)
		webhooks.DELETE("/:id", controller.Delete)
		webhooks.GET("/:id/logs", controller.GetLogs)
	}
}
