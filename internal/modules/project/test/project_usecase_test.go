package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	securityPkg "github.com/Roisfaozi/go-clean-boilerplate/pkg"
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

func TestProjectUseCase_Create_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	req := model.CreateProjectRequest{
		Name:   "Test Project",
		Domain: "test.com",
	}

	deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == "Test Project" && p.Domain == "test.com" && p.OrganizationID == "org-1" && p.UserID == "user-1"
	})).Return(nil)

	resp, err := uc.CreateProject(context.Background(), "user-1", "org-1", req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test Project", resp.Name)
	assert.Equal(t, "test.com", resp.Domain)
	assert.Equal(t, "active", resp.Status)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Create_RepoError(t *testing.T) {
	deps, uc := setupProjectTest()

	req := model.CreateProjectRequest{
		Name:   "Test Project",
		Domain: "test.com",
	}

	deps.Repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	resp, err := uc.CreateProject(context.Background(), "user-1", "org-1", req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "db error", err.Error())
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Create_Sanitization(t *testing.T) {
	deps, uc := setupProjectTest()

	req := model.CreateProjectRequest{
		Name:   "<script>alert(1)</script>Project",
		Domain: "domain<br>.com",
	}

	sanitizedName := securityPkg.SanitizeString(req.Name)
	sanitizedDomain := securityPkg.SanitizeString(req.Domain)

	deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == sanitizedName && p.Domain == sanitizedDomain
	})).Return(nil)

	resp, err := uc.CreateProject(context.Background(), "user-1", "org-1", req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, sanitizedName, resp.Name)
	assert.Equal(t, sanitizedDomain, resp.Domain)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_GetProjects_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	projects := []*entity.Project{
		{ID: "1", Name: "P1", OrganizationID: "org-1"},
		{ID: "2", Name: "P2", OrganizationID: "org-1"},
	}

	deps.Repo.On("GetByOrgID", mock.Anything, "org-1").Return(projects, nil)

	res, err := uc.GetProjects(context.Background(), "org-1")

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "P1", res[0].Name)
	assert.Equal(t, "P2", res[1].Name)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_GetProjects_RepoError(t *testing.T) {
	deps, uc := setupProjectTest()

	deps.Repo.On("GetByOrgID", mock.Anything, "org-1").Return(nil, errors.New("db error"))

	res, err := uc.GetProjects(context.Background(), "org-1")

	assert.Error(t, err)
	assert.Nil(t, res)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_GetProjectByID_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	project := &entity.Project{ID: "1", Name: "P1"}

	deps.Repo.On("GetByID", mock.Anything, "1").Return(project, nil)

	res, err := uc.GetProjectByID(context.Background(), "1")

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "P1", res.Name)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_GetProjectByID_NotFound(t *testing.T) {
	deps, uc := setupProjectTest()

	deps.Repo.On("GetByID", mock.Anything, "1").Return(nil, exception.ErrNotFound)

	res, err := uc.GetProjectByID(context.Background(), "1")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, exception.ErrNotFound, err)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	project := &entity.Project{ID: "1", Name: "Old Name", Domain: "old.com", Status: "active"}
	req := model.UpdateProjectRequest{
		Name:   "New Name",
		Domain: "new.com",
		Status: "inactive",
	}

	deps.Repo.On("GetByID", mock.Anything, "1").Return(project, nil)
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == "New Name" && p.Domain == "new.com" && p.Status == "inactive"
	})).Return(nil)

	res, err := uc.UpdateProject(context.Background(), "1", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "New Name", res.Name)
	assert.Equal(t, "inactive", res.Status)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_NotFound(t *testing.T) {
	deps, uc := setupProjectTest()

	req := model.UpdateProjectRequest{Name: "New Name"}

	deps.Repo.On("GetByID", mock.Anything, "1").Return(nil, exception.ErrNotFound)

	res, err := uc.UpdateProject(context.Background(), "1", req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, exception.ErrNotFound, err)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_RepoError(t *testing.T) {
	deps, uc := setupProjectTest()

	project := &entity.Project{ID: "1", Name: "Old Name"}
	req := model.UpdateProjectRequest{Name: "New Name"}

	deps.Repo.On("GetByID", mock.Anything, "1").Return(project, nil)
	deps.Repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	res, err := uc.UpdateProject(context.Background(), "1", req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "db error", err.Error())
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Update_Sanitization(t *testing.T) {
	deps, uc := setupProjectTest()

	project := &entity.Project{ID: "1", Name: "Old Name"}
	req := model.UpdateProjectRequest{
		Name: "New <b>Name</b>",
	}

	sanitizedName := securityPkg.SanitizeString(req.Name)

	deps.Repo.On("GetByID", mock.Anything, "1").Return(project, nil)
	deps.Repo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Project) bool {
		return p.Name == sanitizedName
	})).Return(nil)

	res, err := uc.UpdateProject(context.Background(), "1", req)

	assert.NoError(t, err)
	assert.Equal(t, sanitizedName, res.Name)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Delete_Success(t *testing.T) {
	deps, uc := setupProjectTest()

	deps.Repo.On("Delete", mock.Anything, "1").Return(nil)

	err := uc.DeleteProject(context.Background(), "1")

	assert.NoError(t, err)
	deps.Repo.AssertExpectations(t)
}

func TestProjectUseCase_Delete_RepoError(t *testing.T) {
	deps, uc := setupProjectTest()

	deps.Repo.On("Delete", mock.Anything, "1").Return(errors.New("db error"))

	err := uc.DeleteProject(context.Background(), "1")

	assert.Error(t, err)
	deps.Repo.AssertExpectations(t)
}
