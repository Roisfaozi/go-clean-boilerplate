// Package database provides database utilities including multi-tenancy context and SQL scopes.
package database

import (
	"context"
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
