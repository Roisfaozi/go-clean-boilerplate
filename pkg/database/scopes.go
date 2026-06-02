// Package database provides database utilities including multi-tenancy GORM scopes.
package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// ContextKey is the type for context keys to avoid collisions
type ContextKey string

const (
	// OrganizationIDKey is the context key for organization ID.
	// This is set by the TenantMiddleware after validating user membership.
	OrganizationIDKey ContextKey = "organization_id"

	// IncludeDeletedOrganizationKey allows privileged callers to opt into
	// querying data that belongs to soft-deleted organizations.
	IncludeDeletedOrganizationKey ContextKey = "include_deleted_organizations"
)

// OrganizationScope returns a GORM scope function that filters queries by organization_id.
// This implements Row-Level Security for multi-tenant data isolation.
//
// Usage in repository:
//
//	db.WithContext(ctx).Scopes(database.OrganizationScope(ctx)).Find(&roles)
//
// The scope will:
//   - Add WHERE organization_id = ? when valid org_id is in context
//   - Skip filtering when context is empty (Super Admin bypass mode)
//   - Skip filtering when org_id is empty string (fail-safe)
//   - Skip filtering when org_id is wrong type (fail-safe)
func OrganizationScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// Extract organization_id from context
		orgIDValue := ctx.Value(OrganizationIDKey)
		if orgIDValue == nil {
			return db
		}

		orgID, ok := orgIDValue.(string)
		if !ok {
			return db
		}

		// Empty string check - fail-safe to avoid WHERE id = ""
		if orgID == "" {
			return db
		}

		// Apply the organization filter: organization-specific OR global (NULL).
		// Parent organization soft-delete visibility is handled separately by
		// OrganizationVisibilityScope so legacy rows without a backing
		// organizations record remain queryable.
		return db.Where("organization_id = ? OR organization_id IS NULL", orgID)
	}
}

// OrganizationVisibilityScope ensures a child resource only remains visible when its
// parent organization is still active, unless the caller is explicitly allowed to
// inspect soft-deleted organizations.
func OrganizationVisibilityScope(ctx context.Context, orgColumn string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if orgColumn == "" || CanAccessDeletedOrganizations(ctx) {
			return db
		}

		return db.Where(
			fmt.Sprintf(
				"(%s IS NULL OR NOT EXISTS (SELECT 1 FROM organizations WHERE organizations.id = %s AND organizations.deleted_at IS NOT NULL AND organizations.deleted_at <> 0))",
				orgColumn,
				orgColumn,
			),
		)
	}
}

// SetOrganizationContext returns a new context with the organization_id set.
// This is used by the TenantMiddleware to inject the org_id into request context.
func SetOrganizationContext(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, OrganizationIDKey, orgID)
}

// SetAllowDeletedOrganizations marks the context as allowed to inspect data that
// belongs to soft-deleted organizations. This should be reserved for privileged
// administrative flows such as superadmin investigation or restore.
func SetAllowDeletedOrganizations(ctx context.Context, allow bool) context.Context {
	return context.WithValue(ctx, IncludeDeletedOrganizationKey, allow)
}

// GetOrganizationID extracts the organization_id from context.
// Returns empty string if not present or wrong type.
func GetOrganizationID(ctx context.Context) string {
	orgIDValue := ctx.Value(OrganizationIDKey)
	if orgIDValue == nil {
		return ""
	}

	orgID, ok := orgIDValue.(string)
	if !ok {
		return ""
	}

	return orgID
}

// CanAccessDeletedOrganizations reports whether the current request context is
// allowed to include data belonging to soft-deleted organizations.
func CanAccessDeletedOrganizations(ctx context.Context) bool {
	value := ctx.Value(IncludeDeletedOrganizationKey)
	allowed, ok := value.(bool)
	return ok && allowed
}
