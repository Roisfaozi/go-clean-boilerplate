package http

import "github.com/gin-gonic/gin"

func RegisterAuthorizedRoutes(router *gin.RouterGroup, handler *AuditHandler) {
	auditGroup := router.Group("/audit-logs")
	{
		auditGroup.POST("/search", handler.GetLogsDynamic)
	}
}
