package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterWebhookRoutes(r *gin.RouterGroup, controller *WebhookController) {
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("", controller.Create)
		webhooks.GET("", controller.FindByOrganization)
		webhooks.GET("/:id", controller.FindByID)
		webhooks.PUT("/:id", controller.Update)
		webhooks.DELETE("/:id", controller.Delete)
		webhooks.GET("/:id/logs", controller.GetLogs)
	}
}
