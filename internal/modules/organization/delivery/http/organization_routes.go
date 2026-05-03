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
		orgGroup.GET("/:id", controller.GetOrganization)
		orgGroup.GET("/slug/:slug", controller.GetOrganizationBySlug)
		orgGroup.PUT("/:id", controller.UpdateOrganization)
		orgGroup.DELETE("/:id", controller.DeleteOrganization)

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

			// Get online members (Presence)
			orgGroup.GET("/:id/presence", controller.GetPresence)
		}
	}
}
