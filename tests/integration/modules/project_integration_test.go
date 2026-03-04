//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	projectEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProjectIntegration(env *setup.TestEnvironment) usecase.ProjectUseCase {
	repo := repository.NewProjectRepository(env.DB)
	return usecase.NewProjectUseCase(repo)
}

func TestProjectIntegration_Create_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupProjectIntegration(env)
	ctx := context.Background()

	req := model.CreateProjectRequest{
		Name:   "Integration Test Project",
		Domain: "integration.example.com",
	}

	result, err := uc.CreateProject(ctx, "user-1", "org-1", req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "Integration Test Project", result.Name)
	assert.Equal(t, "integration.example.com", result.Domain)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "org-1", result.OrganizationID)
	assert.Equal(t, "user-1", result.UserID)
}

func TestProjectIntegration_CRUD_Lifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupProjectIntegration(env)
	ctx := context.Background()
	orgID := "org-lifecycle-" + uuid.NewString()[:8]
	userID := "user-lifecycle"

	// 1. Create
	req := model.CreateProjectRequest{
		Name:   "Lifecycle Project",
		Domain: "lifecycle.example.com",
	}
	created, err := uc.CreateProject(ctx, userID, orgID, req)
	require.NoError(t, err)
	require.NotNil(t, created)
	projectID := created.ID

	// 2. Read by ID
	fetched, err := uc.GetProjectByID(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, "Lifecycle Project", fetched.Name)
	assert.Equal(t, "lifecycle.example.com", fetched.Domain)

	// 3. Update
	updateReq := model.UpdateProjectRequest{
		Name:   "Updated Lifecycle",
		Domain: "updated.example.com",
		Status: "inactive",
	}
	updated, err := uc.UpdateProject(ctx, projectID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated Lifecycle", updated.Name)
	assert.Equal(t, "updated.example.com", updated.Domain)
	assert.Equal(t, "inactive", updated.Status)

	// 4. Delete
	err = uc.DeleteProject(ctx, projectID)
	require.NoError(t, err)

	// 5. Verify deleted (soft delete - GetByID should fail)
	_, err = uc.GetProjectByID(ctx, projectID)
	assert.Error(t, err, "Should not find deleted project")
}

func TestProjectIntegration_GetByOrgID(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupProjectIntegration(env)
	ctx := context.Background()
	orgA := "org-a-" + uuid.NewString()[:8]
	orgB := "org-b-" + uuid.NewString()[:8]

	// Create 2 projects in Org A
	uc.CreateProject(ctx, "user-1", orgA, model.CreateProjectRequest{Name: "Org A Proj 1", Domain: "a1.com"})
	uc.CreateProject(ctx, "user-1", orgA, model.CreateProjectRequest{Name: "Org A Proj 2", Domain: "a2.com"})

	// Create 1 project in Org B
	uc.CreateProject(ctx, "user-2", orgB, model.CreateProjectRequest{Name: "Org B Proj 1", Domain: "b1.com"})

	// Fetch Org A projects
	projectsA, err := uc.GetProjects(ctx, orgA)
	require.NoError(t, err)
	assert.Len(t, projectsA, 2)

	// Fetch Org B projects
	projectsB, err := uc.GetProjects(ctx, orgB)
	require.NoError(t, err)
	assert.Len(t, projectsB, 1)

	// Fetch nonexistent org
	projectsNone, err := uc.GetProjects(ctx, "org-nonexistent")
	require.NoError(t, err)
	assert.Len(t, projectsNone, 0)
}

func TestProjectIntegration_PartialUpdate(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupProjectIntegration(env)
	ctx := context.Background()

	// Create
	created, err := uc.CreateProject(ctx, "user-1", "org-1", model.CreateProjectRequest{
		Name:   "Partial Update Project",
		Domain: "partial.example.com",
	})
	require.NoError(t, err)

	// Partial update - only name
	updated, err := uc.UpdateProject(ctx, created.ID, model.UpdateProjectRequest{
		Name: "Updated Name Only",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Name Only", updated.Name)
	assert.Equal(t, "partial.example.com", updated.Domain, "Domain should not change")
	assert.Equal(t, "active", updated.Status, "Status should not change")
}

func TestProjectIntegration_Security_SQLInjection(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupProjectIntegration(env)
	ctx := context.Background()

	// Attempt SQL injection via project name
	result, err := uc.CreateProject(ctx, "user-1", "org-1", model.CreateProjectRequest{
		Name:   "'; DROP TABLE projects; --",
		Domain: "sqli.example.com",
	})

	// Should succeed (GORM parameterizes queries)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the projects table still exists
	var count int64
	env.DB.Model(&projectEntity.Project{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1), "Projects table should still exist")
}

func TestProjectIntegration_OrganizationScopeIsolation(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	repo := repository.NewProjectRepository(env.DB)
	ctx := context.Background()

	orgA := "org-iso-a-" + uuid.NewString()[:8]
	orgB := "org-iso-b-" + uuid.NewString()[:8]

	// Create project in Org A directly via repo
	projA := &projectEntity.Project{
		OrganizationID: orgA,
		UserID:         "user-a",
		Name:           "Org A Secret Project",
		Domain:         "secret-a.com",
		Status:         "active",
	}
	err := repo.Create(ctx, projA)
	require.NoError(t, err)

	// Try to get the project using Org B's context scope
	ctxOrgB := database.SetOrganizationContext(ctx, orgB)
	_, err = repo.GetByID(ctxOrgB, projA.ID)
	assert.Error(t, err, "Org B should not be able to access Org A's project")
}
