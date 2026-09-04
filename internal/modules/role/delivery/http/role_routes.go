package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAuthorizedRoutes(router *gin.RouterGroup, roleHandler *RoleController) {
	roleGroup := router.Group("/roles")
	{
		roleGroup.POST("", delivery.AdaptHTTPHandler(roleHandler.HTTPCreate))
		roleGroup.GET("", delivery.AdaptHTTPHandler(roleHandler.HTTPGetAll))
		roleGroup.PUT("/:id", delivery.AdaptHTTPHandler(roleHandler.HTTPUpdate))
		roleGroup.POST("/search", delivery.AdaptHTTPHandler(roleHandler.HTTPGetRolesDynamic))
		roleGroup.DELETE("/:id", delivery.AdaptHTTPHandler(roleHandler.HTTPDelete))
	}
}

// RegisterTenantRoutes registers organization-scoped custom role management routes.
func RegisterTenantRoutes(router *gin.RouterGroup, roleHandler *RoleController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	orgRoleGroup := router.Group("/organizations/:id/roles")
	{
		orgRoleGroup.POST("", apiKeyMiddleware.RequireScopes("role:manage"), delivery.AdaptHTTPHandler(roleHandler.HTTPCreateOrganizationRole))
		orgRoleGroup.GET("", apiKeyMiddleware.RequireScopes("role:view", "role:manage"), delivery.AdaptHTTPHandler(roleHandler.HTTPGetOrganizationRoles))
		orgRoleGroup.PUT("/:roleId", apiKeyMiddleware.RequireScopes("role:manage"), delivery.AdaptHTTPHandler(roleHandler.HTTPUpdateOrganizationRole))
		orgRoleGroup.DELETE("/:roleId", apiKeyMiddleware.RequireScopes("role:manage"), delivery.AdaptHTTPHandler(roleHandler.HTTPDeleteOrganizationRole))
	}
}
