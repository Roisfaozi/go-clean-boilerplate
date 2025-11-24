package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterPermissionRoutes registers the permission related HTTP routes.
//
// RegisterPermissionRoutes sets up the routes for assigning roles, granting
// permissions, retrieving permissions, updating permissions, and revoking
// permissions. It takes a *gin.RouterGroup as the first argument and a
// *PermissionHandler as the second argument. The *gin.RouterGroup is used to
// add routes to a specific group of routes, and the *PermissionHandler is used
// to handle the requests to those routes.
//
// The routes registered by this function are:
//   - POST /permissions/assign-role: assigns a role to a user
//   - POST /permissions/grant: grants a permission to a role
//   - GET /permissions: retrieves all permissions
//   - GET /permissions/:role: retrieves permissions for a specific role
//   - PUT /permissions: updates a permission
//   - DELETE /permissions/revoke: revokes a permission from a role
func RegisterPermissionRoutes(router *gin.RouterGroup, handler *PermissionHandler) {
	permissionGroup := router.Group("/permissions")
	{
		permissionGroup.POST("/assign-role", handler.AssignRole)
		permissionGroup.POST("/grant", handler.GrantPermission)
		permissionGroup.GET("", handler.GetAllPermissions)
		permissionGroup.GET("/:role", handler.GetPermissionsForRole)
		permissionGroup.PUT("", handler.UpdatePermission)
		permissionGroup.DELETE("/revoke", handler.RevokePermission)
	}
}
