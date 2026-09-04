package database

import (
	"context"
)

// SQLOrganizationClause returns the WHERE SQL fragment and args for organization isolation.
// If orgID is empty, it returns no clause (superadmin mode / bypass).
// Otherwise, it filters by organization_id = ? OR organization_id IS NULL.
func SQLOrganizationClause(ctx context.Context, col string) (string, []any) {
	orgID := GetOrganizationID(ctx)
	if orgID == "" {
		return "", nil
	}
	if col == "" {
		col = "organization_id"
	}
	return "(" + col + " = ? OR " + col + " IS NULL)", []any{orgID}
}

// SQLOrganizationVisibilityClause returns the SQL fragment ensuring parent organization is active.
func SQLOrganizationVisibilityClause(ctx context.Context, orgCol string) (string, []any) {
	if orgCol == "" || CanAccessDeletedOrganizations(ctx) {
		return "", nil
	}
	sql := "(" + orgCol + " IS NULL OR NOT EXISTS (SELECT 1 FROM organizations WHERE organizations.id = " + orgCol + " AND organizations.deleted_at IS NOT NULL AND organizations.deleted_at <> 0))"
	return sql, nil
}
