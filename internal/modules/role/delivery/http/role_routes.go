package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterAuthorizedRoutes registers role routes that require authentication.
func RegisterAuthorizedRoutes(router *gin.RouterGroup, roleHandler *RoleHandler) {
	roleGroup := router.Group("/roles")
	// Ensure this group is protected by an admin-only authorization middleware in the main router setup.
	{
		roleGroup.POST("", roleHandler.Create)
		roleGroup.GET("", roleHandler.GetAll)
	}
}