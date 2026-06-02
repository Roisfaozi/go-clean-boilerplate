package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type scopeTestOrganization struct {
	ID        string `gorm:"column:id;primaryKey"`
	DeletedAt *int64 `gorm:"column:deleted_at"`
}

func (scopeTestOrganization) TableName() string {
	return "organizations"
}

type scopeTestResource struct {
	ID             string `gorm:"column:id;primaryKey"`
	OrganizationID string `gorm:"column:organization_id"`
}

func (scopeTestResource) TableName() string {
	return "scope_test_resources"
}

func TestOrganizationVisibilityScope_AllowsActiveAndLegacyOrganizations(t *testing.T) {
	uid, _ := uuid.NewV7()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uid.String())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(&scopeTestOrganization{}, &scopeTestResource{}))

	deletedAt := int64(1710000000000)
	require.NoError(t, db.Create(&scopeTestOrganization{ID: "org-null"}).Error)
	require.NoError(t, db.Create(&scopeTestOrganization{ID: "org-zero", DeletedAt: ptrInt64(0)}).Error)
	require.NoError(t, db.Create(&scopeTestOrganization{ID: "org-deleted", DeletedAt: &deletedAt}).Error)

	require.NoError(t, db.Create(&scopeTestResource{ID: "res-null", OrganizationID: "org-null"}).Error)
	require.NoError(t, db.Create(&scopeTestResource{ID: "res-zero", OrganizationID: "org-zero"}).Error)
	require.NoError(t, db.Create(&scopeTestResource{ID: "res-deleted", OrganizationID: "org-deleted"}).Error)
	require.NoError(t, db.Create(&scopeTestResource{ID: "res-orphan", OrganizationID: "org-missing"}).Error)

	var visible []scopeTestResource
	err = db.WithContext(context.Background()).
		Scopes(OrganizationVisibilityScope(context.Background(), "scope_test_resources.organization_id")).
		Order("id ASC").
		Find(&visible).Error
	require.NoError(t, err)

	assert.Len(t, visible, 3)
	assert.Equal(t, []string{"res-null", "res-orphan", "res-zero"}, []string{visible[0].ID, visible[1].ID, visible[2].ID})
}

func ptrInt64(v int64) *int64 {
	return &v
}

// NOTE: Full GORM integration tests for OrganizationScope are in:
// tests/integration/modules/organization_integration_test.go
// These tests use real MySQL via testcontainers to verify:
// - Scope injection (WHERE organization_id = ?)
// - Data isolation between tenants
// - Super Admin bypass mode
