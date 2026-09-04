package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
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
		permissionGroup.POST("/assign-role", delivery.AdaptHTTPHandler(controller.HTTPAssignRole))
		permissionGroup.DELETE("/revoke-role", delivery.AdaptHTTPHandler(controller.HTTPRevokeRole))
		permissionGroup.POST("/grant", delivery.AdaptHTTPHandler(controller.HTTPGrantPermission))
		permissionGroup.GET("", delivery.AdaptHTTPHandler(controller.HTTPGetAllPermissions))
		permissionGroup.GET("/:role", delivery.AdaptHTTPHandler(controller.HTTPGetPermissionsForRole))
		permissionGroup.GET("/roles/:role/users", delivery.AdaptHTTPHandler(controller.HTTPGetUsersForRole))
		permissionGroup.PUT("", delivery.AdaptHTTPHandler(controller.HTTPUpdatePermission))
		permissionGroup.DELETE("/revoke", delivery.AdaptHTTPHandler(controller.HTTPRevokePermission))

		permissionGroup.POST("/inheritance", delivery.AdaptHTTPHandler(controller.HTTPAddRoleInheritance))
		permissionGroup.DELETE("/inheritance", delivery.AdaptHTTPHandler(controller.HTTPRemoveRoleInheritance))
		permissionGroup.GET("/:role/parents", delivery.AdaptHTTPHandler(controller.HTTPGetParentRoles))

		// New routes for Matrix View
		permissionGroup.GET("/resources", delivery.AdaptHTTPHandler(controller.HTTPGetResourceAggregation))
		permissionGroup.GET("/inheritance-tree", delivery.AdaptHTTPHandler(controller.HTTPGetInheritanceTree))

		// Bulk Access Right assignment
		permissionGroup.GET("/roles/:role/access-rights", delivery.AdaptHTTPHandler(controller.HTTPGetRoleAccessRights))
		permissionGroup.POST("/assign-access-right", delivery.AdaptHTTPHandler(controller.HTTPAssignAccessRight))
		permissionGroup.DELETE("/revoke-access-right", delivery.AdaptHTTPHandler(controller.HTTPRevokeAccessRight))
	}
}

// RegisterBatchCheckRoute registers the route for batch permission checking which requires authentication but not specific admin permissions.
func RegisterBatchCheckRoute(router *gin.RouterGroup, controller *PermissionController) {
	router.POST("/permissions/check-batch", delivery.AdaptHTTPHandler(controller.HTTPBatchCheck))
}
