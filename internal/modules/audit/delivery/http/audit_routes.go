package http

import "github.com/gin-gonic/gin"

func RegisterAuthorizedRoutes(router *gin.RouterGroup, handler *AuditController) {
	auditGroup := router.Group("/audit-logs")
	{
		auditGroup.POST("/search", handler.GetLogsDynamic)
	}
}
