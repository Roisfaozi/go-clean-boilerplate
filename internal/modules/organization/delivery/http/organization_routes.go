package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
)

var (
	OrganizationBaseRoutes = "/organizations"
	
)

func RegisterAuthenticatedRoutes(router *gin.RouterGroup, controller *OrganizationController, idempotencyMiddleware gin.HandlerFunc) {
	orgGroup := router.Group(OrganizationBaseRoutes)
	{
		// Create new organization
		if idempotencyMiddleware != nil {
			orgGroup.POST("", idempotencyMiddleware, controller.CreateOrganization)
		} else {
			orgGroup.POST("", controller.CreateOrganization)
		}

		// Get organizations the user is a member of
		orgGroup.GET("/me", controller.GetMyOrganizations)
	}
}

// RegisterPublicRoutes registers routes that do not require authentication or tenant context
func RegisterPublicRoutes(router *gin.RouterGroup, controller *OrganizationController, idempotencyMiddleware gin.HandlerFunc) {
	orgGroup := router.Group(OrganizationBaseRoutes)
	{
		if idempotencyMiddleware != nil {
			orgGroup.POST("/invitations/accept", idempotencyMiddleware, controller.AcceptInvitation)
		} else {
			orgGroup.POST("/invitations/accept", controller.AcceptInvitation)
		}
	}
}

// RegisterTenantRoutes registers routes that require tenant context
// These routes use TenantMiddleware to set organization context
func RegisterTenantRoutes(router *gin.RouterGroup, controller *OrganizationController, apiKeyMiddleware *middleware.APIKeyMiddleware, idempotencyMiddleware gin.HandlerFunc) {
	orgGroup := router.Group(OrganizationBaseRoutes)
	{
		orgGroup.GET("/:id", apiKeyMiddleware.RequireScopes("org:view", "org:manage"), controller.GetOrganization)
		orgGroup.GET("/slug/:slug", apiKeyMiddleware.RequireScopes("org:view", "org:manage"), controller.GetOrganizationBySlug)
		orgGroup.PUT("/:id", apiKeyMiddleware.RequireScopes("org:manage"), controller.UpdateOrganization)
		orgGroup.DELETE("/:id", apiKeyMiddleware.RequireUserSession(), controller.DeleteOrganization)

		// Member Management
		membersGroup := orgGroup.Group("/:id/members")
		{
			// Invite member
			if idempotencyMiddleware != nil {
				membersGroup.POST("/invite", idempotencyMiddleware, apiKeyMiddleware.RequireScopes("member:manage"), controller.InviteMember)
			} else {
				membersGroup.POST("/invite", apiKeyMiddleware.RequireScopes("member:manage"), controller.InviteMember)
			}

			// Get all members
			membersGroup.GET("", apiKeyMiddleware.RequireScopes("member:manage"), controller.GetMembers)

			// Update member role
			membersGroup.PATCH("/:userId", apiKeyMiddleware.RequireScopes("member:manage"), controller.UpdateMemberRole)

			// Remove member
			membersGroup.DELETE("/:userId", apiKeyMiddleware.RequireScopes("member:manage"), controller.RemoveMember)

			// Get online members (Presence)
			orgGroup.GET("/:id/presence", apiKeyMiddleware.RequireScopes("presence:view"), controller.GetPresence)
		}
	}
}

// RegisterAdminRoutes registers privileged organization routes intended for superadmin flows.
func RegisterAdminRoutes(router *gin.RouterGroup, controller *OrganizationController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	orgGroup := router.Group(OrganizationBaseRoutes)
	{
		orgGroup.POST("/:id/restore", apiKeyMiddleware.RequireUserSession(), controller.RestoreOrganization)
		orgGroup.DELETE("/:id/hard", apiKeyMiddleware.RequireUserSession(), controller.HardDeleteOrganization)
	}
}
