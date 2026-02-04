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

// TestOrganizationScope_ReturnsFunction tests that OrganizationScope returns a valid scope function
func TestOrganizationScope_ReturnsFunction(t *testing.T) {
	ctx := context.WithValue(context.Background(), OrganizationIDKey, "org-789")

	scopeFunc := OrganizationScope(ctx)

	assert.NotNil(t, scopeFunc, "OrganizationScope should return a non-nil function")
}

// NOTE: Full GORM integration tests for OrganizationScope are in:
// tests/integration/modules/organization_integration_test.go
// These tests use real MySQL via testcontainers to verify:
// - Scope injection (WHERE organization_id = ?)
// - Data isolation between tenants
// - Super Admin bypass mode
