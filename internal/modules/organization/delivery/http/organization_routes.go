package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterAuthenticatedRoutes registers organization routes that require authentication
// but NOT organization-level authorization (can access any org data)
func RegisterAuthenticatedRoutes(router *gin.RouterGroup, controller *OrganizationController) {
	orgGroup := router.Group("/organizations")
	{
		// Create new organization
		orgGroup.POST("", controller.CreateOrganization)

		// Get organizations the user is a member of
		orgGroup.GET("/me", controller.GetMyOrganizations)

		// Get organization by ID (requires membership - use middleware in future)
		orgGroup.GET("/:id", controller.GetOrganization)

		// Get organization by slug
		orgGroup.GET("/slug/:slug", controller.GetOrganizationBySlug)

		// Update organization (requires owner/admin role - use middleware)
		orgGroup.PUT("/:id", controller.UpdateOrganization)

		// Delete organization (owner only)
		orgGroup.DELETE("/:id", controller.DeleteOrganization)
	}
}

// RegisterPublicRoutes registers routes that do not require authentication or tenant context
func RegisterPublicRoutes(router *gin.RouterGroup, controller *OrganizationController) {
	orgGroup := router.Group("/organizations")
	{
		orgGroup.POST("/invitations/accept", controller.AcceptInvitation)
	}
}

// RegisterTenantRoutes registers routes that require tenant context
// These routes use TenantMiddleware to set organization context
func RegisterTenantRoutes(router *gin.RouterGroup, controller *OrganizationController) {
	orgGroup := router.Group("/organizations")
	{
		// Member Management
		membersGroup := orgGroup.Group("/:id/members")
		{
			// Invite member
			membersGroup.POST("/invite", controller.InviteMember)

			// Get all members
			membersGroup.GET("", controller.GetMembers)

			// Update member role
			membersGroup.PATCH("/:userId", controller.UpdateMemberRole)

			// Remove member
			membersGroup.DELETE("/:userId", controller.RemoveMember)
		}
	}
}
