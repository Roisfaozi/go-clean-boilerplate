package middleware

import (
	"context"
	"errors"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	// OrgIDHeader is the header name for organization ID
	OrgIDHeader = "X-Organization-ID"
	// OrgSlugHeader is the header name for organization slug
	OrgSlugHeader = "X-Organization-Slug"
)

// TenantMiddleware validates organization membership and sets org context.
// Uses IOrganizationReader for cached membership validation.
type TenantMiddleware struct {
	OrgRepo repository.OrganizationRepository
	Reader  usecase.IOrganizationReader
	Log     *logrus.Logger
}

// NewTenantMiddleware creates a new TenantMiddleware instance
func NewTenantMiddleware(
	orgRepo repository.OrganizationRepository,
	reader usecase.IOrganizationReader,
	log *logrus.Logger,
) *TenantMiddleware {
	return &TenantMiddleware{
		OrgRepo: orgRepo,
		Reader:  reader,
		Log:     log,
	}
}

// RequireOrganization validates that the request includes a valid organization
// and that the authenticated user is an active member
func (m *TenantMiddleware) RequireOrganization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from auth context (set by AuthMiddleware)
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			response.Unauthorized(c, errors.New("user not authenticated"), "unauthorized")
			c.Abort()
			return
		}

		// Get organization ID from header (preferred), or route parameter if explicitly named :orgID
		// CRITICAL FIX: Do NOT use c.Param("id") loosely as it might be resource ID (e.g. /projects/:id)
		orgID := c.GetHeader(OrgIDHeader)
		if orgID == "" {
			orgID = c.Param("orgId") // Strict naming convention
		}

		// Fallback: If not found in header or specific param, check query param
		if orgID == "" {
			orgID = c.Query("orgId")
		}

		orgSlug := c.GetHeader(OrgSlugHeader)

		if orgID == "" && orgSlug == "" {
			response.BadRequest(c, errors.New("organization ID or slug is required"), "missing organization identifier")
			c.Abort()
			return
		}

		// Resolve org ID from slug if needed
		if orgID == "" && orgSlug != "" {
			org, err := m.OrgRepo.FindBySlug(c.Request.Context(), orgSlug)
			if err != nil {
				m.Log.WithError(err).Error("Failed to lookup organization by slug")
				response.InternalServerError(c, err, "internal server error")
				c.Abort()
				return
			}
			if org == nil {
				response.NotFound(c, errors.New("organization not found"), "organization not found")
				c.Abort()
				return
			}
			orgID = org.ID
		}

		// Check membership using cached reader
		isMember, err := m.Reader.ValidateMembership(c.Request.Context(), orgID, userID)
		if err != nil {
			m.Log.WithError(err).Error("Failed to validate membership")
			response.InternalServerError(c, err, "internal server error")
			c.Abort()
			return
		}

		if !isMember {
			response.Forbidden(c, errors.New("user is not a member of this organization"), "access denied")
			c.Abort()
			return
		}

		// Get member role for context
		role, err := m.Reader.GetMemberRole(c.Request.Context(), orgID, userID)
		if err != nil {
			m.Log.WithError(err).Warn("Failed to get member role, proceeding without role context")
		}

		// Set organization context for downstream handlers and repository scopes
		ctx := database.SetOrganizationContext(c.Request.Context(), orgID)
		c.Request = c.Request.WithContext(ctx)

		// Set organization info in Gin context for easy access
		c.Set("organization_id", orgID)
		c.Set("member_role", role)

		c.Next()
	}
}

// OptionalOrganization extracts organization context if provided but does not require it.
// Useful for routes that work with or without organization context.
func (m *TenantMiddleware) OptionalOrganization() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			// No authenticated user, skip organization context
			c.Next()
			return
		}

		orgID := c.GetHeader(OrgIDHeader)
		if orgID == "" {
			orgID = c.Param("orgId")
		}
		if orgID == "" {
			orgID = c.Query("orgId")
		}

		orgSlug := c.GetHeader(OrgSlugHeader)

		if orgID == "" && orgSlug == "" {
			// No org specified, proceed without org context
			c.Next()
			return
		}

		// Resolve org ID from slug if needed
		if orgID == "" && orgSlug != "" {
			org, err := m.OrgRepo.FindBySlug(c.Request.Context(), orgSlug)
			if err != nil || org == nil {
				// Org not found, proceed without org context
				c.Next()
				return
			}
			orgID = org.ID
		}

		// Validate membership
		isMember, err := m.Reader.ValidateMembership(c.Request.Context(), orgID, userID)
		if err != nil || !isMember {
			// Not a member, proceed without org context
			c.Next()
			return
		}

		// Get member role
		role, _ := m.Reader.GetMemberRole(c.Request.Context(), orgID, userID)

		// Set organization context
		ctx := database.SetOrganizationContext(c.Request.Context(), orgID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("organization_id", orgID)
		c.Set("member_role", role)

		c.Next()
	}
}

// RequireOrgRole validates that the user has a specific role (or higher) in the organization.
// Must be used after RequireOrganization middleware.
func (m *TenantMiddleware) RequireOrgRole(allowedRoles ...string) gin.HandlerFunc {
	roleHierarchy := map[string]int{
		"owner":  3,
		"admin":  2,
		"member": 1,
	}

	return func(c *gin.Context) {
		role, exists := GetMemberRoleFromContext(c)
		if !exists {
			response.Forbidden(c, errors.New("organization role not found"), "access denied")
			c.Abort()
			return
		}

		userLevel := roleHierarchy[role]
		for _, allowedRole := range allowedRoles {
			if roleHierarchy[allowedRole] <= userLevel {
				c.Next()
				return
			}
		}

		response.Forbidden(c, errors.New("insufficient permissions"), "access denied")
		c.Abort()
	}
}

// GetOrganizationIDFromContext extracts organization ID from Gin context
func GetOrganizationIDFromContext(c *gin.Context) (string, bool) {
	orgID, exists := c.Get("organization_id")
	if !exists {
		return "", false
	}
	orgIDStr, ok := orgID.(string)
	if !ok || orgIDStr == "" {
		return "", false
	}
	return orgIDStr, true
}

// GetMemberRoleFromContext extracts member's role from Gin context
func GetMemberRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("member_role")
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	if !ok || roleStr == "" {
		return "", false
	}
	return roleStr, true
}

// InvalidateMembershipCache delegates cache invalidation to the reader
func (m *TenantMiddleware) InvalidateMembershipCache(ctx context.Context, orgID, userID string) error {
	return m.Reader.InvalidateMembershipCache(ctx, orgID, userID)
}
