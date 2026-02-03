package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// OrgIDHeader is the header name for organization ID
	OrgIDHeader = "X-Organization-ID"
	// OrgSlugHeader is the header name for organization slug
	OrgSlugHeader = "X-Organization-Slug"
	// MembershipCacheTTL is the TTL for membership cache in Redis
	MembershipCacheTTL = 5 * time.Minute
	// MembershipCachePrefix is the Redis key prefix for membership cache
	MembershipCachePrefix = "membership:"
)

// MembershipCache represents cached membership data
type MembershipCache struct {
	OrgID  string `json:"org_id"`
	RoleID string `json:"role_id"`
	Status string `json:"status"`
}

// TenantMiddleware validates organization membership and sets org context
type TenantMiddleware struct {
	OrgRepo       repository.OrganizationRepository
	MemberRepo    repository.OrganizationMemberRepository
	Redis         *redis.Client
	Log           *logrus.Logger
	EnableCache   bool
}

// NewTenantMiddleware creates a new TenantMiddleware instance
func NewTenantMiddleware(
	orgRepo repository.OrganizationRepository,
	memberRepo repository.OrganizationMemberRepository,
	redisClient *redis.Client,
	log *logrus.Logger,
) *TenantMiddleware {
	return &TenantMiddleware{
		OrgRepo:     orgRepo,
		MemberRepo:  memberRepo,
		Redis:       redisClient,
		Log:         log,
		EnableCache: redisClient != nil,
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

		// Get organization ID from header (preferred) or slug
		orgID := c.GetHeader(OrgIDHeader)
		orgSlug := c.GetHeader(OrgSlugHeader)

		if orgID == "" && orgSlug == "" {
			response.BadRequest(c, errors.New("organization ID or slug is required"), "missing organization header")
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

		// Check membership (with Redis cache)
		membership, err := m.checkMembership(c.Request.Context(), orgID, userID)
		if err != nil {
			m.Log.WithError(err).Error("Failed to check membership")
			response.InternalServerError(c, err, "internal server error")
			c.Abort()
			return
		}

		if membership == nil {
			response.Forbidden(c, errors.New("user is not a member of this organization"), "access denied")
			c.Abort()
			return
		}

		// Check if member is active
		if membership.Status != "active" {
			response.Forbidden(c, fmt.Errorf("membership status is %s", membership.Status), "access denied")
			c.Abort()
			return
		}

		// Set organization context for downstream handlers and repository scopes
		ctx := database.SetOrganizationContext(c.Request.Context(), orgID)
		c.Request = c.Request.WithContext(ctx)

		// Set organization info in Gin context for easy access
		c.Set("organization_id", orgID)
		c.Set("member_role_id", membership.RoleID)
		c.Set("member_status", membership.Status)

		c.Next()
	}
}

// checkMembership checks if a user is a member of an organization
// Uses Redis cache if available
func (m *TenantMiddleware) checkMembership(ctx context.Context, orgID, userID string) (*MembershipCache, error) {
	cacheKey := fmt.Sprintf("%s%s:%s", MembershipCachePrefix, orgID, userID)

	// Try cache first
	if m.EnableCache && m.Redis != nil {
		cached, err := m.Redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var membership MembershipCache
			if json.Unmarshal([]byte(cached), &membership) == nil {
				return &membership, nil
			}
		}
		// Cache miss or error - continue to DB lookup
	}

	// Lookup from database
	status, err := m.MemberRepo.GetMemberStatus(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	if status == "" {
		// Not a member - cache negative result
		if m.EnableCache && m.Redis != nil {
			// Cache "not a member" for shorter duration
			_ = m.Redis.Set(ctx, cacheKey, "null", time.Minute).Err()
		}
		return nil, nil
	}

	// Build membership cache
	membership := &MembershipCache{
		OrgID:  orgID,
		Status: status,
		// Note: RoleID would need additional lookup - simplified for now
	}

	// Cache the result
	if m.EnableCache && m.Redis != nil {
		if data, err := json.Marshal(membership); err == nil {
			_ = m.Redis.Set(ctx, cacheKey, data, MembershipCacheTTL).Err()
		}
	}

	return membership, nil
}

// InvalidateMembershipCache invalidates the membership cache for a user in an org
func (m *TenantMiddleware) InvalidateMembershipCache(ctx context.Context, orgID, userID string) error {
	if !m.EnableCache || m.Redis == nil {
		return nil
	}
	cacheKey := fmt.Sprintf("%s%s:%s", MembershipCachePrefix, orgID, userID)
	return m.Redis.Del(ctx, cacheKey).Err()
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

// GetMemberRoleIDFromContext extracts member's role ID from Gin context
func GetMemberRoleIDFromContext(c *gin.Context) (string, bool) {
	roleID, exists := c.Get("member_role_id")
	if !exists {
		return "", false
	}
	roleIDStr, ok := roleID.(string)
	if !ok || roleIDStr == "" {
		return "", false
	}
	return roleIDStr, true
}
