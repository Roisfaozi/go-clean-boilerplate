package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterPermissionRoutes registers the permission related HTTP routes.
//
// RegisterPermissionRoutes sets up the routes for assigning roles, granting
// permissions, retrieving permissions, updating permissions, and revoking
// permissions. It takes a *gin.RouterGroup as the first argument and a
// *PermissionController as the second argument. The *gin.RouterGroup is used to
// add routes to a specific group of routes, and the *PermissionController is used
// to handle the requests to those routes.
//
// The routes registered by this function are:
//   - POST /permissions/assign-role: assigns a role to a user
//   - POST /permissions/grant: grants a permission to a role
//   - GET /permissions: retrieves all permissions
//   - GET /permissions/:role: retrieves permissions for a specific role
//   - PUT /permissions: updates a permission
//   - DELETE /permissions/revoke: revokes a permission from a role
func RegisterPermissionRoutes(router *gin.RouterGroup, controller *PermissionController) {
	permissionGroup := router.Group("/permissions")
	{
		permissionGroup.POST("/assign-role", controller.AssignRole)
		permissionGroup.POST("/grant", controller.GrantPermission)
		permissionGroup.GET("", controller.GetAllPermissions)
		permissionGroup.GET("/:role", controller.GetPermissionsForRole)
		permissionGroup.PUT("", controller.UpdatePermission)
		permissionGroup.DELETE("/revoke", controller.RevokePermission)
	}
}