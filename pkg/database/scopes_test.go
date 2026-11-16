package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetOrganizationID tests the helper function to extract org_id from context
func TestGetOrganizationID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "Valid org_id in context",
			ctx:      context.WithValue(context.Background(), OrganizationIDKey, "org-123"),
			expected: "org-123",
		},
		{
			name:     "Empty context",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "Empty string org_id",
			ctx:      context.WithValue(context.Background(), OrganizationIDKey, ""),
			expected: "",
		},
		{
			name:     "Wrong type in context (int)",
			ctx:      context.WithValue(context.Background(), OrganizationIDKey, 12345),
			expected: "",
		},
		{
			name:     "Wrong type in context (nil)",
			ctx:      context.WithValue(context.Background(), OrganizationIDKey, nil),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOrganizationID(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSetOrganizationContext tests the helper function to set org_id in context
func TestSetOrganizationContext(t *testing.T) {
	ctx := context.Background()
	orgID := "org-456"

	newCtx := SetOrganizationContext(ctx, orgID)

	// Verify the value was set
	result := GetOrganizationID(newCtx)
	assert.Equal(t, orgID, result)

	// Verify original context was not modified
	original := GetOrganizationID(ctx)
	assert.Equal(t, "", original)
}

func TestCanAccessDeletedOrganizations(t *testing.T) {
	ctx := context.Background()
	assert.False(t, CanAccessDeletedOrganizations(ctx))

	allowedCtx := SetAllowDeletedOrganizations(ctx, true)
	assert.True(t, CanAccessDeletedOrganizations(allowedCtx))

	disallowedCtx := SetAllowDeletedOrganizations(ctx, false)
	assert.False(t, CanAccessDeletedOrganizations(disallowedCtx))
}

func TestSQLOrganizationClause(t *testing.T) {
	clause, args := SQLOrganizationClause(context.Background(), "")
	assert.Equal(t, "", clause)
	assert.Nil(t, args)

	ctx := SetOrganizationContext(context.Background(), "org-789")
	clause, args = SQLOrganizationClause(ctx, "")
	assert.Equal(t, "(organization_id = ? OR organization_id IS NULL)", clause)
	assert.Equal(t, []any{"org-789"}, args)

	clause, args = SQLOrganizationClause(ctx, "r.organization_id")
	assert.Equal(t, "(r.organization_id = ? OR r.organization_id IS NULL)", clause)
	assert.Equal(t, []any{"org-789"}, args)
}

func TestSQLOrganizationVisibilityClause(t *testing.T) {
	clause, args := SQLOrganizationVisibilityClause(context.Background(), "")
	assert.Equal(t, "", clause)
	assert.Nil(t, args)

	clause, args = SQLOrganizationVisibilityClause(context.Background(), "r.organization_id")
	assert.Contains(t, clause, "r.organization_id IS NULL OR NOT EXISTS")
	assert.Nil(t, args)

	allowedCtx := SetAllowDeletedOrganizations(context.Background(), true)
	clause, args = SQLOrganizationVisibilityClause(allowedCtx, "r.organization_id")
	assert.Equal(t, "", clause)
	assert.Nil(t, args)
}
