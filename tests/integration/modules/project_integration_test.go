//go:build integration
// +build integration

package modules

import (
	"context"
	"fmt"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProjectDependencies(env *setup.TestEnvironment) (usecase.ProjectUseCase, repository.ProjectRepository) {
	repo := repository.NewProjectRepository(env.DB)
	return usecase.NewProjectUseCase(repo), repo
}

func TestProjectIntegration_CRUD_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	projectUC, projectRepo := setupProjectDependencies(env)

	userID := "user-123"
	orgID := "org-123"

	// 1. Create
	createReq := model.CreateProjectRequest{
		Name:   "Integration Project",
		Domain: "integration.example.com",
	}

	created, err := projectUC.CreateProject(context.Background(), userID, orgID, createReq)
	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, createReq.Name, created.Name)
	assert.Equal(t, createReq.Domain, created.Domain)
	assert.Equal(t, userID, created.UserID)
	assert.Equal(t, orgID, created.OrganizationID)
	assert.NotEmpty(t, created.ID)

	// 2. Get By ID
	fetched, err := projectUC.GetProjectByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.Name, fetched.Name)

	// 3. Update
	updateReq := model.UpdateProjectRequest{
		Name: "Updated Project Name",
	}
	updated, err := projectUC.UpdateProject(context.Background(), created.ID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated Project Name", updated.Name)
	assert.Equal(t, created.Domain, updated.Domain) // Should remain unchanged

	// Verify in DB
	dbProject, err := projectRepo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Project Name", dbProject.Name)

	// 4. Get Projects List
	list, err := projectUC.GetProjects(context.Background(), orgID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, updated.ID, list[0].ID)

	// 5. Delete
	err = projectUC.DeleteProject(context.Background(), created.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = projectUC.GetProjectByID(context.Background(), created.ID)
	assert.Error(t, err)
	assert.Equal(t, exception.ErrNotFound, err)
}

func TestProjectIntegration_Security_XSS(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	projectUC, projectRepo := setupProjectDependencies(env)

	userID := "user-xss"
	orgID := "org-xss"

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert(1)>",
		"javascript:alert('1')", // Added quotes to force escape
	}

	for i, payload := range xssPayloads {
		t.Run(fmt.Sprintf("Payload_%d", i), func(t *testing.T) {
			req := model.CreateProjectRequest{
				Name:   payload,
				Domain: payload,
			}

			created, err := projectUC.CreateProject(context.Background(), userID, orgID, req)
			require.NoError(t, err)

			// Verify response is sanitized (depends on pkg.SanitizeString implementation)
			// pkg.SanitizeString uses html.EscapeString
			expected := pkg.SanitizeString(payload)

			assert.Equal(t, expected, created.Name)
			assert.Equal(t, expected, created.Domain)

			// Verify DB storage is sanitized
			dbProject, err := projectRepo.GetByID(context.Background(), created.ID)
			require.NoError(t, err)
			assert.Equal(t, expected, dbProject.Name)
			assert.Equal(t, expected, dbProject.Domain)

			// Verify it is NOT the original payload
			assert.NotEqual(t, payload, dbProject.Name)
		})
	}
}

func TestProjectIntegration_Security_UpdateXSS(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	projectUC, projectRepo := setupProjectDependencies(env)

	userID := "user-update-xss"
	orgID := "org-update-xss"

	// Create valid project
	created, err := projectUC.CreateProject(context.Background(), userID, orgID, model.CreateProjectRequest{
		Name:   "Safe Name",
		Domain: "safe.com",
	})
	require.NoError(t, err)

	// Update with XSS
	payload := "<script>alert('update')</script>"
	updateReq := model.UpdateProjectRequest{
		Name: payload,
	}

	updated, err := projectUC.UpdateProject(context.Background(), created.ID, updateReq)
	require.NoError(t, err)

	expected := pkg.SanitizeString(payload)
	assert.Equal(t, expected, updated.Name)

	dbProject, err := projectRepo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, expected, dbProject.Name)
}
