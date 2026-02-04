package modules

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	orgA := "org-a-" + time.Now().Format("20060102150405")
	orgB := "org-b-" + time.Now().Format("20060102150405")

	// Create Users
	userA := &entity.User{
		ID:             "user-a-iso",
		Username:       "usera_iso",
		Email:          "user.a.iso@example.com",
		Password:       "password",
		Name:           "User A Iso",
		OrganizationID: &orgA,
		Status:         entity.UserStatusActive,
		CreatedAt:      time.Now().UnixMilli(),
	}

	userB := &entity.User{
		ID:             "user-b-iso",
		Username:       "userb_iso",
		Email:          "user.b.iso@example.com",
		Password:       "password",
		Name:           "User B Iso",
		OrganizationID: &orgB,
		Status:         entity.UserStatusActive,
		CreatedAt:      time.Now().UnixMilli(),
	}

	ctx := context.Background()
	// Create directly using repo (which has no scope on Create)
	require.NoError(t, repo.Create(ctx, userA))
	require.NoError(t, repo.Create(ctx, userB))

	// Test 1: Scope to Org A
	ctxOrgA := database.SetOrganizationContext(ctx, orgA)
	usersA, _, err := repo.FindAll(ctxOrgA, &model.GetUserListRequest{})
	assert.NoError(t, err)
	
	foundA := false
	foundB := false
	for _, u := range usersA {
		if u.ID == userA.ID { foundA = true }
		if u.ID == userB.ID { foundB = true }
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
		if u.ID == userA.ID { foundA = true }
		if u.ID == userB.ID { foundB = true }
	}
	assert.False(t, foundA, "Should NOT find user A in Org B context")
	assert.True(t, foundB, "Should find user B in Org B context")

	// Test 3: GetByOrganization Explicit
	explicitUsers, err := repo.GetByOrganization(ctx, orgA)
	assert.NoError(t, err)
	assert.NotEmpty(t, explicitUsers)
	assert.Equal(t, userA.ID, explicitUsers[0].ID)
}
