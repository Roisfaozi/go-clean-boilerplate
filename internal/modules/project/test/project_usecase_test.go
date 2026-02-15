package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type projectTestDeps struct {
	Repo *mocks.MockProjectRepository
}

func setupProjectTest() (*projectTestDeps, usecase.ProjectUseCase) {
	deps := &projectTestDeps{
		Repo: new(mocks.MockProjectRepository),
	}
	uc := usecase.NewProjectUseCase(deps.Repo)
	return deps, uc
}

// === CreateProject Tests ===

func TestProjectUseCase_Create_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	deps.Repo.On("Create", ctx, mock.AnythingOfType("*entity.Project")).Return(nil).Once()

	req := model.CreateProjectRequest{Name: "My Project", Domain: "myproject.com"}
	result, err := uc.CreateProject(ctx, "user-1", "org-1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "My Project", result.Name)
	assert.Equal(t, "myproject.com", result.Domain)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, "org-1", result.OrganizationID)
	assert.Equal(t, "user-1", result.UserID)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Create_RepositoryError(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	repoErr := errors.New("db connection failed")
	deps.Repo.On("Create", ctx, mock.AnythingOfType("*entity.Project")).Return(repoErr).Once()

	req := model.CreateProjectRequest{Name: "Fail Project", Domain: "fail.com"}
	result, err := uc.CreateProject(ctx, "user-1", "org-1", req)

	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	assert.Nil(t, result)
	deps.Repo.AssertExpectations(t)
}

// === GetProjects Tests ===

func TestProjectUseCase_GetProjects_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	expectedProjects := []*entity.Project{
		{ID: "p1", OrganizationID: "org-1", UserID: "u1", Name: "Project 1", Domain: "p1.com", Status: "active"},
		{ID: "p2", OrganizationID: "org-1", UserID: "u2", Name: "Project 2", Domain: "p2.com", Status: "active"},
	}
	deps.Repo.On("GetByOrgID", ctx, "org-1").Return(expectedProjects, nil).Once()

	results, err := uc.GetProjects(ctx, "org-1")

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Project 1", results[0].Name)
	assert.Equal(t, "Project 2", results[1].Name)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_GetProjects_Empty(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	deps.Repo.On("GetByOrgID", ctx, "org-empty").Return([]*entity.Project{}, nil).Once()

	results, err := uc.GetProjects(ctx, "org-empty")

	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 0)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_GetProjects_RepositoryError(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	repoErr := errors.New("db error")
	deps.Repo.On("GetByOrgID", ctx, "org-1").Return(nil, repoErr).Once()

	results, err := uc.GetProjects(ctx, "org-1")

	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	assert.Nil(t, results)
	deps.Repo.AssertExpectations(t)
}

// === GetProjectByID Tests ===

func TestProjectUseCase_GetByID_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	expected := &entity.Project{
		ID: "p1", OrganizationID: "org-1", UserID: "u1",
		Name: "Found Project", Domain: "found.com", Status: "active",
		CreatedAt: 1000, UpdatedAt: 2000,
	}
	deps.Repo.On("GetByID", ctx, "p1").Return(expected, nil).Once()

	result, err := uc.GetProjectByID(ctx, "p1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "p1", result.ID)
	assert.Equal(t, "Found Project", result.Name)
	assert.Equal(t, int64(1000), result.CreatedAt)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_GetByID_NotFound(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	deps.Repo.On("GetByID", ctx, "nonexistent").Return(nil, gorm.ErrRecordNotFound).Once()

	result, err := uc.GetProjectByID(ctx, "nonexistent")

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrNotFound)
	assert.Nil(t, result)
	deps.Repo.AssertExpectations(t)
}

// === UpdateProject Tests ===

func TestProjectUseCase_Update_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	existing := &entity.Project{
		ID: "p1", OrganizationID: "org-1", UserID: "u1",
		Name: "Old Name", Domain: "old.com", Status: "active",
	}
	deps.Repo.On("GetByID", ctx, "p1").Return(existing, nil).Once()
	deps.Repo.On("Update", ctx, mock.AnythingOfType("*entity.Project")).Return(nil).Once()

	req := model.UpdateProjectRequest{Name: "New Name", Domain: "new.com", Status: "inactive"}
	result, err := uc.UpdateProject(ctx, "p1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Name", result.Name)
	assert.Equal(t, "new.com", result.Domain)
	assert.Equal(t, "inactive", result.Status)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_PartialUpdate_NameOnly(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	existing := &entity.Project{
		ID: "p1", OrganizationID: "org-1", UserID: "u1",
		Name: "Old Name", Domain: "keep.com", Status: "active",
	}
	deps.Repo.On("GetByID", ctx, "p1").Return(existing, nil).Once()
	deps.Repo.On("Update", ctx, mock.AnythingOfType("*entity.Project")).Return(nil).Once()

	// Only update name, leave domain and status unchanged
	req := model.UpdateProjectRequest{Name: "Updated Name"}
	result, err := uc.UpdateProject(ctx, "p1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Name", result.Name)
	assert.Equal(t, "keep.com", result.Domain, "Domain should remain unchanged")
	assert.Equal(t, "active", result.Status, "Status should remain unchanged")
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_NotFound(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	deps.Repo.On("GetByID", ctx, "nonexistent").Return(nil, gorm.ErrRecordNotFound).Once()

	req := model.UpdateProjectRequest{Name: "Updated"}
	result, err := uc.UpdateProject(ctx, "nonexistent", req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, exception.ErrNotFound)
	assert.Nil(t, result)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_RepositoryError(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	existing := &entity.Project{
		ID: "p1", OrganizationID: "org-1", UserID: "u1",
		Name: "Old Name", Domain: "old.com", Status: "active",
	}
	deps.Repo.On("GetByID", ctx, "p1").Return(existing, nil).Once()

	repoErr := errors.New("update failed")
	deps.Repo.On("Update", ctx, mock.AnythingOfType("*entity.Project")).Return(repoErr).Once()

	req := model.UpdateProjectRequest{Name: "New Name"}
	result, err := uc.UpdateProject(ctx, "p1", req)

	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	assert.Nil(t, result)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_EmptyRequest(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	existing := &entity.Project{
		ID: "p1", OrganizationID: "org-1", UserID: "u1",
		Name: "Keep Name", Domain: "keep.com", Status: "active",
	}
	deps.Repo.On("GetByID", ctx, "p1").Return(existing, nil).Once()
	deps.Repo.On("Update", ctx, mock.AnythingOfType("*entity.Project")).Return(nil).Once()

	// Empty update request - should keep all existing values
	req := model.UpdateProjectRequest{}
	result, err := uc.UpdateProject(ctx, "p1", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Keep Name", result.Name)
	assert.Equal(t, "keep.com", result.Domain)
	assert.Equal(t, "active", result.Status)
	deps.Repo.AssertExpectations(t)
}

// === DeleteProject Tests ===

func TestProjectUseCase_Delete_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	deps.Repo.On("Delete", ctx, "p1").Return(nil).Once()

	err := uc.DeleteProject(ctx, "p1")

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Delete_RepositoryError(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	repoErr := errors.New("delete failed")
	deps.Repo.On("Delete", ctx, "p1").Return(repoErr).Once()

	err := uc.DeleteProject(ctx, "p1")

	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	deps.Repo.AssertExpectations(t)
}
