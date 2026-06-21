// Package database provides database utilities including multi-tenancy GORM scopes.
package database

import (
	"context"

	"gorm.io/gorm"
)

// ContextKey is the type for context keys to avoid collisions
type ContextKey string

const (
	// OrganizationIDKey is the context key for organization ID.
	// This is set by the TenantMiddleware after validating user membership.
	OrganizationIDKey ContextKey = "organization_id"
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
			// No org_id in context - Super Admin mode or global routes
			return db
		}

		// Type assertion - must be string
		orgID, ok := orgIDValue.(string)
		if !ok {
			// Wrong type in context - fail-safe, don't apply filter
			return db
		}

		// Empty string check - fail-safe to avoid WHERE id = ""
		if orgID == "" {
			return db
		}

		// Apply the organization filter
		return db.Where("organization_id = ?", orgID)
	}
}

// SetOrganizationContext returns a new context with the organization_id set.
// This is used by the TenantMiddleware to inject the org_id into request context.
func SetOrganizationContext(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, OrganizationIDKey, orgID)
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
