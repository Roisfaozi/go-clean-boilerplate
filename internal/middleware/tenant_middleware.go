package middleware

import (
	"context"
	"errors"
	"strings"

	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
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
	OrgSlugHeader  = "X-Organization-Slug"
	superAdminRole = "role:superadmin"
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

		allowDeleted := canAccessDeletedOrganizations(c)
		orgID, requested := requestedOrganizationID(c)
		orgSlug := requestedOrganizationSlug(c)
		if !requested && orgSlug == "" {
			response.BadRequest(c, errors.New("organization ID or slug is required"), "missing organization identifier")
			c.Abort()
			return
		}

		org, orgID, err := m.resolveOrganization(c, orgID, orgSlug, allowDeleted)
		if err != nil {
			m.respondOrganizationLookupError(c, err)
			return
		}

		role := ""
		if allowDeleted {
			role = superAdminRole
		} else {
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
			role, err = m.Reader.GetMemberRole(c.Request.Context(), orgID, userID)
			if err != nil {
				m.Log.WithError(err).Warn("Failed to get member role, proceeding without role context")
			}
		}

		m.applyOrganizationContext(c, orgID, role, allowDeleted, org != nil && org.DeletedAt != 0)

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

		allowDeleted := canAccessDeletedOrganizations(c)
		orgID, requested := requestedOrganizationID(c)
		orgSlug := requestedOrganizationSlug(c)
		if !requested && orgSlug == "" {
			// No org specified, proceed without org context
			c.Next()
			return
		}

		org, orgID, err := m.resolveOrganization(c, orgID, orgSlug, allowDeleted)
		if err != nil {
			m.respondOrganizationLookupError(c, err)
			return
		}

		role := ""
		if allowDeleted {
			role = superAdminRole
		} else {
			// Validate membership
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

			// Get member role
			role, _ = m.Reader.GetMemberRole(c.Request.Context(), orgID, userID)
		}

		m.applyOrganizationContext(c, orgID, role, allowDeleted, org != nil && org.DeletedAt != 0)

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

func requestedOrganizationID(c *gin.Context) (string, bool) {
	if orgID := c.GetHeader(OrgIDHeader); orgID != "" {
		return orgID, true
	}
	if orgID, exists := GetOrganizationIDFromContext(c); exists {
		return orgID, true
	}
	if orgID := requestedOrganizationRouteID(c); orgID != "" {
		return orgID, true
	}
	return "", false
}

func requestedOrganizationSlug(c *gin.Context) string {
	if orgSlug := c.GetHeader(OrgSlugHeader); orgSlug != "" {
		return orgSlug
	}
	return c.Param("slug")
}

func canAccessDeletedOrganizations(c *gin.Context) bool {
	role, ok := GetRoleFromContext(c)
	return ok && role == superAdminRole
}

func requestedOrganizationRouteID(c *gin.Context) string {
	fullPath := c.FullPath()
	if fullPath == "" {
		fullPath = c.Request.URL.Path
	}

	if strings.HasPrefix(fullPath, "/api/v1/organizations/:id") {
		return c.Param("id")
	}

	return ""
}

func (m *TenantMiddleware) resolveOrganization(c *gin.Context, orgID, orgSlug string, allowDeleted bool) (*orgEntity.Organization, string, error) {
	ctx := c.Request.Context()
	if allowDeleted {
		ctx = database.SetAllowDeletedOrganizations(ctx, true)
	}

	if orgID != "" {
		org, err := m.OrgRepo.FindByID(ctx, orgID)
		if err != nil {
			return nil, "", err
		}
		if org == nil {
			return nil, "", errors.New("organization not found")
		}
		return org, org.ID, nil
	}

	org, err := m.OrgRepo.FindBySlug(ctx, orgSlug)
	if err != nil {
		return nil, "", err
	}
	if org == nil {
		return nil, "", errors.New("organization not found")
	}
	return org, org.ID, nil
}

func (m *TenantMiddleware) respondOrganizationLookupError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if err.Error() == "organization not found" {
		response.NotFound(c, err, "organization not found")
		c.Abort()
		return
	}

	m.Log.WithError(err).Error("Failed to resolve organization")
	response.InternalServerError(c, err, "internal server error")
	c.Abort()
}

func (m *TenantMiddleware) applyOrganizationContext(c *gin.Context, orgID, role string, allowDeleted, isDeleted bool) {
	ctx := database.SetOrganizationContext(c.Request.Context(), orgID)
	if allowDeleted {
		ctx = database.SetAllowDeletedOrganizations(ctx, true)
	}

	c.Request = c.Request.WithContext(ctx)
	c.Set("organization_id", orgID)
	if role != "" {
		c.Set("member_role", role)
	}
	if isDeleted {
		c.Set("organization_deleted", true)
	}
}
