package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
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
			orgGroup.POST("", idempotencyMiddleware, delivery.AdaptHTTPHandler(controller.HTTPCreateOrganization))
		} else {
			orgGroup.POST("", delivery.AdaptHTTPHandler(controller.HTTPCreateOrganization))
		}

		// Get organizations the user is a member of
		orgGroup.GET("/me", delivery.AdaptHTTPHandler(controller.HTTPGetMyOrganizations))
	}
}

// RegisterPublicRoutes registers routes that do not require authentication or tenant context
func RegisterPublicRoutes(router *gin.RouterGroup, controller *OrganizationController, idempotencyMiddleware gin.HandlerFunc) {
	orgGroup := router.Group(OrganizationBaseRoutes)
	{
		if idempotencyMiddleware != nil {
			orgGroup.POST("/invitations/accept", idempotencyMiddleware, delivery.AdaptHTTPHandler(controller.HTTPAcceptInvitation))
		} else {
			orgGroup.POST("/invitations/accept", delivery.AdaptHTTPHandler(controller.HTTPAcceptInvitation))
		}
	}
}

// RegisterTenantRoutes registers routes that require tenant context
// These routes use TenantMiddleware to set organization context
func RegisterTenantRoutes(router *gin.RouterGroup, controller *OrganizationController, apiKeyMiddleware *middleware.APIKeyMiddleware, idempotencyMiddleware gin.HandlerFunc) {
	orgGroup := router.Group(OrganizationBaseRoutes)
	{
		orgGroup.GET("/:id", apiKeyMiddleware.RequireScopes("org:view", "org:manage"), delivery.AdaptHTTPHandler(controller.HTTPGetOrganization))
		orgGroup.GET("/slug/:slug", apiKeyMiddleware.RequireScopes("org:view", "org:manage"), delivery.AdaptHTTPHandler(controller.HTTPGetOrganizationBySlug))
		orgGroup.PUT("/:id", apiKeyMiddleware.RequireScopes("org:manage"), delivery.AdaptHTTPHandler(controller.HTTPUpdateOrganization))
		orgGroup.DELETE("/:id", apiKeyMiddleware.RequireUserSession(), delivery.AdaptHTTPHandler(controller.HTTPDeleteOrganization))

		// Member Management
		membersGroup := orgGroup.Group("/:id/members")
		{
			// Invite member
			if idempotencyMiddleware != nil {
				membersGroup.POST("/invite", idempotencyMiddleware, apiKeyMiddleware.RequireScopes("member:manage"), delivery.AdaptHTTPHandler(controller.HTTPInviteMember))
			} else {
				membersGroup.POST("/invite", apiKeyMiddleware.RequireScopes("member:manage"), delivery.AdaptHTTPHandler(controller.HTTPInviteMember))
			}

			// Get all members
			membersGroup.GET("", apiKeyMiddleware.RequireScopes("member:manage"), delivery.AdaptHTTPHandler(controller.HTTPGetMembers))

			// Update member role
			membersGroup.PATCH("/:userId", apiKeyMiddleware.RequireScopes("member:manage"), delivery.AdaptHTTPHandler(controller.HTTPUpdateMemberRole))

			// Remove member
			membersGroup.DELETE("/:userId", apiKeyMiddleware.RequireScopes("member:manage"), delivery.AdaptHTTPHandler(controller.HTTPRemoveMember))

			// Get online members (Presence)
			orgGroup.GET("/:id/presence", apiKeyMiddleware.RequireScopes("presence:view"), delivery.AdaptHTTPHandler(controller.HTTPGetPresence))
		}
	}
}

// RegisterAdminRoutes registers privileged organization routes intended for superadmin flows.
func RegisterAdminRoutes(router *gin.RouterGroup, controller *OrganizationController, apiKeyMiddleware *middleware.APIKeyMiddleware) {
	orgGroup := router.Group(OrganizationBaseRoutes)
	{
		orgGroup.POST("/:id/restore", apiKeyMiddleware.RequireUserSession(), delivery.AdaptHTTPHandler(controller.HTTPRestoreOrganization))
		orgGroup.DELETE("/:id/hard", apiKeyMiddleware.RequireUserSession(), delivery.AdaptHTTPHandler(controller.HTTPHardDeleteOrganization))
	}
}
