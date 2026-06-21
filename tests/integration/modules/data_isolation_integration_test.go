package modules

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
)

func TestDataIsolation_User_FindAll(t *testing.T) {
	// Use the correct Setup function from 'setup' package
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return // Setup skipped
	}
	// Defer cleanup if needed, though SetupIntegrationEnvironment reuses containers via singleton

	repo := repository.NewUserRepository(env.DB, env.Logger)

	// Setup Headers
	orgA := "org-a"
	orgB := "org-b"

	// Create Organizations in DB first to satisfy foreign key if any, and for consistency
	env.DB.Exec("DELETE FROM organization_members")
	env.DB.Exec("DELETE FROM organizations")
	env.DB.Exec("INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		orgA, "Org A", "org-a", "system", "active", time.Now().UnixMilli(), time.Now().UnixMilli())
	env.DB.Exec("INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		orgB, "Org B", "org-b", "system", "active", time.Now().UnixMilli(), time.Now().UnixMilli())

	// Create Users with memberships
	userA := setup.CreateTestUser(t, env.DB, "usera_iso", "user.a.iso@example.com", "password", orgA)
	userB := setup.CreateTestUser(t, env.DB, "userb_iso", "user.b.iso@example.com", "password", orgB)

	// Verify memberships exist
	var count int64
	env.DB.Table("organization_members").Where("organization_id = ?", orgA).Count(&count)
	assert.Equal(t, int64(1), count, "Org A should have 1 member")
	env.DB.Table("organization_members").Where("organization_id = ?", orgB).Count(&count)
	assert.Equal(t, int64(1), count, "Org B should have 1 member")

	ctx := context.Background()

	// Test 1: Scope to Org A
	ctxOrgA := database.SetOrganizationContext(ctx, orgA)
	usersA, _, err := repo.FindAll(ctxOrgA, &model.GetUserListRequest{})
	assert.NoError(t, err)

	foundA := false
	foundB := false
	for _, u := range usersA {
		if u.ID == userA.ID {
			foundA = true
		}
		if u.ID == userB.ID {
			foundB = true
		}
	}
	assert.True(t, foundA, "Should find user A in Org A context")
	assert.False(t, foundB, "Should NOT find user B in Org A context")

	// Test 2: Scope to Org B
	ctxOrgB := database.SetOrganizationContext(ctx, orgB)
	usersB, _, err := repo.FindAll(ctxOrgB, &model.GetUserListRequest{})
	assert.NoError(t, err)

	foundA = false
	foundB = false
	for _, u := range usersB {
		if u.ID == userA.ID {
			foundA = true
		}
		if u.ID == userB.ID {
			foundB = true
		}
	}
	assert.False(t, foundA, "Should NOT find user A in Org B context")
	assert.True(t, foundB, "Should find user B in Org B context")

	// Test 3: GetByOrganization Explicit
	explicitUsers, err := repo.GetByOrganization(ctx, orgA)
	assert.NoError(t, err)
	assert.NotEmpty(t, explicitUsers)
	assert.Equal(t, userA.ID, explicitUsers[0].ID)
}
