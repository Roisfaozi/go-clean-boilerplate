package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAuthorizedRoutes(router *gin.RouterGroup, roleHandler *RoleController) {
	roleGroup := router.Group("/roles")
	{
		roleGroup.POST("", roleHandler.Create)
		roleGroup.GET("", roleHandler.GetAll)
		roleGroup.PUT("/:id", roleHandler.Update)
		roleGroup.POST("/search", roleHandler.GetRolesDynamic)
		roleGroup.DELETE("/:id", roleHandler.Delete)
	}
}

// RegisterTenantRoutes registers organization-scoped custom role management routes.
func RegisterTenantRoutes(router *gin.RouterGroup, roleHandler *RoleController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	orgRoleGroup := router.Group("/organizations/:id/roles")
	{
		orgRoleGroup.POST("", apiKeyMiddleware.RequireScopes("role:manage"), roleHandler.CreateOrganizationRole)
		orgRoleGroup.GET("", apiKeyMiddleware.RequireScopes("role:view", "role:manage"), roleHandler.GetOrganizationRoles)
		orgRoleGroup.PUT("/:roleId", apiKeyMiddleware.RequireScopes("role:manage"), roleHandler.UpdateOrganizationRole)
		orgRoleGroup.DELETE("/:roleId", apiKeyMiddleware.RequireScopes("role:manage"), roleHandler.DeleteOrganizationRole)
	}
}
