package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterPermissionRoutes(router *gin.RouterGroup, handler *PermissionHandler) {
	permissionGroup := router.Group("/permissions")
	{
		permissionGroup.POST("/assign-role", handler.AssignRole)
		permissionGroup.POST("/grant", handler.GrantPermission)
		permissionGroup.GET("", handler.GetAllPermissions)
		permissionGroup.GET("/:role", handler.GetPermissionsForRole)
		permissionGroup.PUT("", handler.UpdatePermission)
	}
}
