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
	repo := new(mocks.MockProjectRepository)
	return &projectTestDeps{Repo: repo}, usecase.NewProjectUseCase(repo)
}

func TestProjectUseCase_CreateProject_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	req := model.CreateProjectRequest{
		Name:   "My Project",
		Domain: "example.com",
	}

	deps.Repo.On("Create", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == req.Name && p.Domain == req.Domain && p.UserID == "user1" && p.OrganizationID == "org1"
	})).Return(nil)

	res, err := uc.CreateProject(ctx, "user1", "org1", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, req.Name, res.Name)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_CreateProject_XSS(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	// Input with XSS payload
	req := model.CreateProjectRequest{
		Name:   "<script>alert('xss')</script>",
		Domain: "javascript:alert(1)",
	}

	// Expect repository to receive SANITIZED input
	// pkg.SanitizeString uses html.EscapeString, so <script> becomes &lt;script&gt;
	expectedName := "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"
	expectedDomain := "javascript:alert(1)"

	deps.Repo.On("Create", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == expectedName && p.Domain == expectedDomain
	})).Return(nil)

	res, err := uc.CreateProject(ctx, "user1", "org1", req)

	assert.NoError(t, err)
	assert.Equal(t, expectedName, res.Name)
	assert.Equal(t, expectedDomain, res.Domain)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_GetProjectByID_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	project := &entity.Project{
		ID:             "proj1",
		OrganizationID: "org1",
		Name:           "Project 1",
		Status:         "active",
		CreatedAt:      time.Now().UnixMilli(),
	}

	deps.Repo.On("GetByID", ctx, "proj1").Return(project, nil)

	res, err := uc.GetProjectByID(ctx, "proj1")

	assert.NoError(t, err)
	assert.Equal(t, project.ID, res.ID)
}

func TestProjectUseCase_GetProjectByID_NotFound(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	// Simulate Repository returning ErrNotFound (since repository now converts DB error)
	deps.Repo.On("GetByID", ctx, "proj1").Return(nil, exception.ErrNotFound)

	res, err := uc.GetProjectByID(ctx, "proj1")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, exception.ErrNotFound, err)
}

func TestProjectUseCase_GetProjectByID_DBError(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	dbErr := errors.New("db connection failed")
	deps.Repo.On("GetByID", ctx, "proj1").Return(nil, dbErr)

	res, err := uc.GetProjectByID(ctx, "proj1")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, dbErr, err)
}

func TestProjectUseCase_UpdateProject_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	existingProject := &entity.Project{
		ID:   "proj1",
		Name: "Old Name",
	}

	req := model.UpdateProjectRequest{
		Name: "New Name",
	}

	deps.Repo.On("GetByID", ctx, "proj1").Return(existingProject, nil)
	deps.Repo.On("Update", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == "New Name"
	})).Return(nil)

	res, err := uc.UpdateProject(ctx, "proj1", req)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", res.Name)
}

func TestProjectUseCase_UpdateProject_XSS(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	existingProject := &entity.Project{
		ID:   "proj1",
		Name: "Old Name",
	}

	req := model.UpdateProjectRequest{
		Name: "<script>alert(1)</script>",
	}

	expectedName := "&lt;script&gt;alert(1)&lt;/script&gt;"

	deps.Repo.On("GetByID", ctx, "proj1").Return(existingProject, nil)
	deps.Repo.On("Update", ctx, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == expectedName
	})).Return(nil)

	res, err := uc.UpdateProject(ctx, "proj1", req)

	assert.NoError(t, err)
	assert.Equal(t, expectedName, res.Name)
}

func TestProjectUseCase_GetProjects_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	projects := []*entity.Project{
		{ID: "p1", Name: "Project 1"},
		{ID: "p2", Name: "Project 2"},
	}

	deps.Repo.On("GetByOrgID", ctx, "org1").Return(projects, nil)

	res, err := uc.GetProjects(ctx, "org1")

	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestProjectUseCase_DeleteProject_Success(t *testing.T) {
	deps, uc := setupProjectTest()
	ctx := context.Background()

	deps.Repo.On("Delete", ctx, "proj1").Return(nil)

	err := uc.DeleteProject(ctx, "proj1")

	assert.NoError(t, err)
}
