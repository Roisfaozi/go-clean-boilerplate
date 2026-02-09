//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"
	"time"

	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupProjectIntegration(env *setup.TestEnvironment) usecase.ProjectUseCase {
	repo := repository.NewProjectRepository(env.DB)
	return usecase.NewProjectUseCase(repo)
}

func createTestOrganization(t *testing.T, db *gorm.DB, name string, ownerID string) *orgEntity.Organization {
	org := &orgEntity.Organization{
		ID:        uuid.New().String(),
		Name:      name,
		Slug:      uuid.New().String(), // Ensure unique slug
		OwnerID:   ownerID,
		Status:    "active",
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}
	err := db.Create(org).Error
	require.NoError(t, err, "Failed to create test organization")
	return org
}

func TestProjectIntegration_CRUD(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test: environment not available")
	}

	uc := setupProjectIntegration(env)
	ctx := context.Background()

	// Create User and Organization to satisfy FK constraints
	user := setup.CreateTestUser(t, env.DB, "proj-user-1", "proj-user-1@example.com", "password")
	org := createTestOrganization(t, env.DB, "Org For Project", user.ID)

	userID := user.ID
	orgID := org.ID

	// Create Project
	req := model.CreateProjectRequest{
		Name:   "Integration Project",
		Domain: "integration.example.com",
	}

	created, err := uc.CreateProject(ctx, userID, orgID, req)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, req.Name, created.Name)
	assert.Equal(t, orgID, created.OrganizationID)

	// Get Project By ID
	// Test 1: Get without context (Super Admin / No Context) -> Should find it (if repo allows)
	fetched, err := uc.GetProjectByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)

	// Test 2: Get with CORRECT Context -> Should find it
	ctxWithOrg := database.SetOrganizationContext(ctx, orgID)
	fetchedWithCtx, err := uc.GetProjectByID(ctxWithOrg, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetchedWithCtx.ID)

	// Test 3: Get with WRONG Context -> Should NOT find it
	ctxWithWrongOrg := database.SetOrganizationContext(ctx, "wrong-org")
	_, err = uc.GetProjectByID(ctxWithWrongOrg, created.ID)
	assert.Error(t, err) // Should be Not Found or specific error

	// Get Projects (List)
	// Test List with CORRECT Context
	list, err := uc.GetProjects(ctxWithOrg, orgID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1)

	// Update Project
	updateReq := model.UpdateProjectRequest{
		Name: "Updated Integration Project",
	}
	updated, err := uc.UpdateProject(ctx, created.ID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, updateReq.Name, updated.Name)

	// Delete Project
	err = uc.DeleteProject(ctx, created.ID)
	require.NoError(t, err)

	// Verify Deletion
	_, err = uc.GetProjectByID(ctx, created.ID)
	assert.Error(t, err)
}

func TestProjectIntegration_Isolation(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test: environment not available")
	}

	uc := setupProjectIntegration(env)
	ctx := context.Background()

	// Create User and Orgs
	user := setup.CreateTestUser(t, env.DB, "iso-user", "iso-user@example.com", "password")
	org1 := createTestOrganization(t, env.DB, "Org Iso 1", user.ID)
	org2 := createTestOrganization(t, env.DB, "Org Iso 2", user.ID)

	// Create Project in Org 1
	proj1, err := uc.CreateProject(ctx, user.ID, org1.ID, model.CreateProjectRequest{Name: "P1", Domain: "d1"})
	require.NoError(t, err)

	// Create Project in Org 2
	proj2, err := uc.CreateProject(ctx, user.ID, org2.ID, model.CreateProjectRequest{Name: "P2", Domain: "d2"})
	require.NoError(t, err)

	// Query with Org 1 Context -> Should only see P1
	ctxOrg1 := database.SetOrganizationContext(ctx, org1.ID)

	// Can I access P2 with Org1 context?
	_, err = uc.GetProjectByID(ctxOrg1, proj2.ID)
	assert.Error(t, err, "Should not be able to access project from another org")

	// Can I access P1 with Org1 context?
	res, err := uc.GetProjectByID(ctxOrg1, proj1.ID)
	assert.NoError(t, err)
	assert.Equal(t, proj1.ID, res.ID)

	// Query with Org 2 Context
	ctxOrg2 := database.SetOrganizationContext(ctx, org2.ID)

	// Can I access P1 with Org2 context?
	_, err = uc.GetProjectByID(ctxOrg2, proj1.ID)
	assert.Error(t, err, "Should not be able to access project from another org")
}
