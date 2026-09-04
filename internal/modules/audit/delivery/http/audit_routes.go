package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/gin-gonic/gin"
)

func RegisterAuthorizedRoutes(router *gin.RouterGroup, controller *AuditController) {
	auditGroup := router.Group("/audit-logs")
	{
		auditGroup.POST("/search", delivery.AdaptHTTPHandler(controller.HTTPGetLogsDynamic))
		auditGroup.GET("/export", delivery.AdaptHTTPHandler(controller.HTTPExport))
		auditGroup.GET("/export-async", delivery.AdaptHTTPHandler(controller.HTTPExportAsync))
	}
}
