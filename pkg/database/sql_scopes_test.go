package database_test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestSQLOrganizationClause(t *testing.T) {
	ctx := context.Background()

	// Empty context
	clause, args := database.SQLOrganizationClause(ctx, "organization_id")
	assert.Empty(t, clause)
	assert.Nil(t, args)

	// Context with org ID
	ctx = database.SetOrganizationContext(ctx, "org-123")
	clause, args = database.SQLOrganizationClause(ctx, "organization_id")
	assert.Equal(t, "(organization_id = ? OR organization_id IS NULL)", clause)
	assert.Equal(t, []any{"org-123"}, args)

	// Custom column
	clause, args = database.SQLOrganizationClause(ctx, "projects.organization_id")
	assert.Equal(t, "(projects.organization_id = ? OR projects.organization_id IS NULL)", clause)
	assert.Equal(t, []any{"org-123"}, args)
}

func TestSQLOrganizationVisibilityClause(t *testing.T) {
	ctx := context.Background()

	clause, args := database.SQLOrganizationVisibilityClause(ctx, "projects.organization_id")
	assert.Contains(t, clause, "NOT EXISTS (SELECT 1 FROM organizations")
	assert.Nil(t, args)

	// With allow deleted
	ctx = database.SetAllowDeletedOrganizations(ctx, true)
	clause, args = database.SQLOrganizationVisibilityClause(ctx, "projects.organization_id")
	assert.Empty(t, clause)
	assert.Nil(t, args)
}
