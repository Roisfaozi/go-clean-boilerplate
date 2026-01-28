package http

import "github.com/gin-gonic/gin"

func RegisterAuthorizedRoutes(router *gin.RouterGroup, controller *AuditController) {
	auditGroup := router.Group("/audit-logs")
	{
		auditGroup.POST("/search", controller.GetLogsDynamic)
		auditGroup.GET("/export", controller.Export)
	}
}