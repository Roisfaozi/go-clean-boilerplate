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
	owner := setup.CreateTestUser(t, env.DB, "project_owner_create", "project_owner_create@test.com", "Password123!")
	org := setup.CreateTestOrganization(t, env.DB, owner.ID, "Project Create Org", "project-create-org-"+uuid.NewString()[:8])

	req := model.CreateProjectRequest{
		Name:   "Integration Test Project",
		Domain: "integration.example.com",
	}

	result, err := uc.CreateProject(ctx, owner.ID, org.ID, req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "Integration Test Project", result.Name)
	assert.Equal(t, "integration.example.com", result.Domain)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, org.ID, result.OrganizationID)
	assert.Equal(t, owner.ID, result.UserID)
}

func TestProjectIntegration_CRUD_Lifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupProjectIntegration(env)
	ctx := context.Background()
	user := setup.CreateTestUser(t, env.DB, "project_lifecycle", "project_lifecycle@test.com", "Password123!")
	org := setup.CreateTestOrganization(t, env.DB, user.ID, "Project Lifecycle Org", "project-lifecycle-org-"+uuid.NewString()[:8])
	orgID := org.ID
	userID := user.ID

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
	name := "Updated Lifecycle"
	domain := "updated.example.com"
	status := "inactive"
	updateReq := model.UpdateProjectRequest{
		Name:   &name,
		Domain: &domain,
		Status: &status,
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
	userA := setup.CreateTestUser(t, env.DB, "project_org_a", "project_org_a@test.com", "Password123!")
	userB := setup.CreateTestUser(t, env.DB, "project_org_b", "project_org_b@test.com", "Password123!")
	orgA := setup.CreateTestOrganization(t, env.DB, userA.ID, "Project Org A", "project-org-a-"+uuid.NewString()[:8]).ID
	orgB := setup.CreateTestOrganization(t, env.DB, userB.ID, "Project Org B", "project-org-b-"+uuid.NewString()[:8]).ID

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
	user := setup.CreateTestUser(t, env.DB, "project_partial", "project_partial@test.com", "Password123!")
	org := setup.CreateTestOrganization(t, env.DB, user.ID, "Project Partial Org", "project-partial-org-"+uuid.NewString()[:8])

	// Create
	created, err := uc.CreateProject(ctx, user.ID, org.ID, model.CreateProjectRequest{
		Name:   "Partial Update Project",
		Domain: "partial.example.com",
	})
	require.NoError(t, err)

	// Partial update - only name
	nameStr := "Updated Name Only"
	updated, err := uc.UpdateProject(ctx, created.ID, model.UpdateProjectRequest{
		Name: &nameStr,
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
	user := setup.CreateTestUser(t, env.DB, "project_sqli", "project_sqli@test.com", "Password123!")
	org := setup.CreateTestOrganization(t, env.DB, user.ID, "Project SQLi Org", "project-sqli-org-"+uuid.NewString()[:8])

	// Attempt SQL injection via project name
	result, err := uc.CreateProject(ctx, user.ID, org.ID, model.CreateProjectRequest{
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

	userA := setup.CreateTestUser(t, env.DB, "project_repo_iso_a", "project_repo_iso_a@test.com", "Password123!")
	userB := setup.CreateTestUser(t, env.DB, "project_repo_iso_b", "project_repo_iso_b@test.com", "Password123!")
	orgA := setup.CreateTestOrganization(t, env.DB, userA.ID, "Project Repo Iso A", "project-repo-iso-a-"+uuid.NewString()[:8]).ID
	orgB := setup.CreateTestOrganization(t, env.DB, userB.ID, "Project Repo Iso B", "project-repo-iso-b-"+uuid.NewString()[:8]).ID

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

func TestProjectIsolation_UseCase(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	repo := repository.NewProjectRepository(env.DB)
	uc := usecase.NewProjectUseCase(repo)
	ctx := context.Background()

	userA := setup.CreateTestUser(t, env.DB, "project_uc_iso_a", "project_uc_iso_a@test.com", "Password123!")
	userB := setup.CreateTestUser(t, env.DB, "project_uc_iso_b", "project_uc_iso_b@test.com", "Password123!")
	orgA := setup.CreateTestOrganization(t, env.DB, userA.ID, "Project UseCase Iso A", "project-uc-iso-a-"+uuid.NewString()[:8]).ID
	orgB := setup.CreateTestOrganization(t, env.DB, userB.ID, "Project UseCase Iso B", "project-uc-iso-b-"+uuid.NewString()[:8]).ID

	// 1. Create project in Org A context
	ctxOrgA := database.SetOrganizationContext(ctx, orgA)
	created, err := uc.CreateProject(ctxOrgA, userA.ID, orgA, model.CreateProjectRequest{
		Name:   "Org A Private Project",
		Domain: "a.example.com",
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	t.Run("Cross-tenant GET should fail", func(t *testing.T) {
		ctxOrgB := database.SetOrganizationContext(ctx, orgB)
		_, err := uc.GetProjectByID(ctxOrgB, created.ID)
		assert.Error(t, err, "Should not be able to fetch project from another organization")
	})

	t.Run("Cross-tenant UPDATE should fail", func(t *testing.T) {
		ctxOrgB := database.SetOrganizationContext(ctx, orgB)
		newName := "Hacked Name"
		_, err := uc.UpdateProject(ctxOrgB, created.ID, model.UpdateProjectRequest{
			Name: &newName,
		})
		assert.Error(t, err, "Should not be able to update project from another organization")

		// Verify name was not changed
		refetched, _ := uc.GetProjectByID(ctxOrgA, created.ID)
		assert.Equal(t, "Org A Private Project", refetched.Name)
	})

	t.Run("Cross-tenant DELETE should fail", func(t *testing.T) {
		ctxOrgB := database.SetOrganizationContext(ctx, orgB)
		err := uc.DeleteProject(ctxOrgB, created.ID)
		// Note: Delete in GORM might not return error if 0 rows affected unless checked,
		// but since we use Scopes, it will just do 'DELETE ... WHERE id = X AND org_id = B'
		// which affects 0 rows.
		assert.NoError(t, err) // It doesn't error, but it shouldn't delete the record

		// Verify project still exists in Org A
		refetched, err := uc.GetProjectByID(ctxOrgA, created.ID)
		assert.NoError(t, err)
		assert.NotNil(t, refetched)
	})
}
