package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestProjectUseCase_CreateProject_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	userID := "user-123"
	orgID := "org-123"
	req := model.CreateProjectRequest{
		Name:   "My Project",
		Domain: "example.com",
	}

	deps.Repo.On("Create", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.OrganizationID == orgID &&
			p.UserID == userID &&
			p.Name == req.Name &&
			p.Domain == req.Domain &&
			p.Status == "active"
	})).Return(nil)

	res, err := uc.CreateProject(ctx, userID, orgID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
	assert.Equal(t, req.Domain, res.Domain)
	assert.Equal(t, "active", res.Status)
	assert.Equal(t, orgID, res.OrganizationID)
	assert.Equal(t, userID, res.UserID)
}

func TestProjectUseCase_CreateProject_Sanitization(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	userID := "user-123"
	orgID := "org-123"
	req := model.CreateProjectRequest{
		Name:   "<script>alert('xss')</script>",
		Domain: "example.com",
	}

	expectedName := "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"

	deps.Repo.On("Create", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == expectedName &&
			p.Domain == req.Domain
	})).Return(nil)

	res, err := uc.CreateProject(ctx, userID, orgID, req)

	assert.NoError(t, err)
	assert.Equal(t, expectedName, res.Name)
}

func TestProjectUseCase_CreateProject_RepoError(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	userID := "user-123"
	orgID := "org-123"
	req := model.CreateProjectRequest{
		Name:   "My Project",
		Domain: "example.com",
	}

	expectedErr := errors.New("db error")
	deps.Repo.On("Create", ctx, mock.Anything).Return(expectedErr)

	res, err := uc.CreateProject(ctx, userID, orgID, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, expectedErr, err)
}

func TestProjectUseCase_GetProjects_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	orgID := "org-123"

	projects := []*entity.Project{
		{
			ID:             "proj-1",
			OrganizationID: orgID,
			Name:           "Project 1",
			Status:         "active",
		},
		{
			ID:             "proj-2",
			OrganizationID: orgID,
			Name:           "Project 2",
			Status:         "active",
		},
	}

	deps.Repo.On("GetByOrgID", ctx, orgID).Return(projects, nil)

	res, err := uc.GetProjects(ctx, orgID)

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "Project 1", res[0].Name)
	assert.Equal(t, "Project 2", res[1].Name)
}

func TestProjectUseCase_GetProjects_RepoError(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	orgID := "org-123"

	expectedErr := errors.New("db error")
	deps.Repo.On("GetByOrgID", ctx, orgID).Return(nil, expectedErr)

	res, err := uc.GetProjects(ctx, orgID)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, expectedErr, err)
}

func TestProjectUseCase_GetProjectByID_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	id := "proj-123"
	project := &entity.Project{
		ID:     id,
		Name:   "My Project",
		Status: "active",
	}

	deps.Repo.On("GetByID", ctx, id).Return(project, nil)

	res, err := uc.GetProjectByID(ctx, id)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, id, res.ID)
	assert.Equal(t, "My Project", res.Name)
}

func TestProjectUseCase_GetProjectByID_NotFound(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	id := "proj-123"

	// Repo usually returns error when record not found
	deps.Repo.On("GetByID", ctx, id).Return(nil, errors.New("record not found"))

	res, err := uc.GetProjectByID(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, exception.ErrNotFound, err)
}

func TestProjectUseCase_UpdateProject_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	id := "proj-123"
	req := model.UpdateProjectRequest{
		Name:   "Updated Project",
		Domain: "updated.com",
		Status: "inactive",
	}

	existingProject := &entity.Project{
		ID:        id,
		Name:      "Old Name",
		Domain:    "old.com",
		Status:    "active",
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}

	deps.Repo.On("GetByID", ctx, id).Return(existingProject, nil)
	deps.Repo.On("Update", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.ID == id &&
			p.Name == req.Name &&
			p.Domain == req.Domain &&
			p.Status == req.Status
	})).Return(nil)

	res, err := uc.UpdateProject(ctx, id, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
	assert.Equal(t, req.Domain, res.Domain)
	assert.Equal(t, req.Status, res.Status)
}

func TestProjectUseCase_UpdateProject_Sanitization(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	id := "proj-123"
	req := model.UpdateProjectRequest{
		Name: "<script>",
	}
	expectedName := "&lt;script&gt;"

	existingProject := &entity.Project{
		ID:   id,
		Name: "Old Name",
	}

	deps.Repo.On("GetByID", ctx, id).Return(existingProject, nil)
	deps.Repo.On("Update", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == expectedName
	})).Return(nil)

	res, err := uc.UpdateProject(ctx, id, req)

	assert.NoError(t, err)
	assert.Equal(t, expectedName, res.Name)
}

func TestProjectUseCase_UpdateProject_NotFound(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	id := "proj-123"
	req := model.UpdateProjectRequest{
		Name: "Updated Project",
	}

	deps.Repo.On("GetByID", ctx, id).Return(nil, errors.New("not found"))

	res, err := uc.UpdateProject(ctx, id, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, exception.ErrNotFound, err)
}

func TestProjectUseCase_UpdateProject_RepoError(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	id := "proj-123"
	req := model.UpdateProjectRequest{
		Name: "Updated Project",
	}

	existingProject := &entity.Project{
		ID:   id,
		Name: "Old Name",
	}

	expectedErr := errors.New("db error")

	deps.Repo.On("GetByID", ctx, id).Return(existingProject, nil)
	deps.Repo.On("Update", ctx, mock.Anything).Return(expectedErr)

	res, err := uc.UpdateProject(ctx, id, req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, expectedErr, err)
}

func TestProjectUseCase_DeleteProject_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	id := "proj-123"

	deps.Repo.On("Delete", ctx, id).Return(nil)

	err := uc.DeleteProject(ctx, id)

	assert.NoError(t, err)
}

func TestProjectUseCase_DeleteProject_RepoError(t *testing.T) {
	deps, uc := setupProjectTest()

	ctx := context.Background()
	id := "proj-123"

	expectedErr := errors.New("db error")
	deps.Repo.On("Delete", ctx, id).Return(expectedErr)

	err := uc.DeleteProject(ctx, id)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
