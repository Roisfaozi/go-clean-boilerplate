package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
)

type ProjectUseCase interface {
	CreateProject(ctx context.Context, userID string, orgID string, req model.CreateProjectRequest) (*model.ProjectResponse, error)
	GetProjects(ctx context.Context, orgID string) ([]*model.ProjectResponse, error)
	GetProjectByID(ctx context.Context, id string) (*model.ProjectResponse, error)
	UpdateProject(ctx context.Context, id string, req model.UpdateProjectRequest) (*model.ProjectResponse, error)
	DeleteProject(ctx context.Context, id string) error
}

type projectUseCase struct {
	repo repository.ProjectRepository
}

func NewProjectUseCase(repo repository.ProjectRepository) ProjectUseCase {
	return &projectUseCase{repo: repo}
}

func (u *projectUseCase) CreateProject(ctx context.Context, userID string, orgID string, req model.CreateProjectRequest) (*model.ProjectResponse, error) {
	req.Name = pkg.SanitizeString(req.Name)
	req.Domain = pkg.SanitizeString(req.Domain)

	project := &entity.Project{
		OrganizationID: orgID,
		UserID:         userID,
		Name:           req.Name,
		Domain:         req.Domain,
		Status:         "active",
	}

	if err := u.repo.Create(ctx, project); err != nil {
		return nil, err
	}

	return u.mapToResponse(project), nil
}

func (u *projectUseCase) GetProjects(ctx context.Context, orgID string) ([]*model.ProjectResponse, error) {
	projects, err := u.repo.GetByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	res := make([]*model.ProjectResponse, len(projects))
	for i, p := range projects {
		res[i] = u.mapToResponse(p)
	}
	return res, nil
}

func (u *projectUseCase) GetProjectByID(ctx context.Context, id string) (*model.ProjectResponse, error) {
	project, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	return u.mapToResponse(project), nil
}

func (u *projectUseCase) UpdateProject(ctx context.Context, id string, req model.UpdateProjectRequest) (*model.ProjectResponse, error) {
	project, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, exception.ErrNotFound
	}

	if req.Name != "" {
		project.Name = pkg.SanitizeString(req.Name)
	}
	if req.Domain != "" {
		project.Domain = pkg.SanitizeString(req.Domain)
	}
	if req.Status != "" {
		project.Status = req.Status
	}

	if err := u.repo.Update(ctx, project); err != nil {
		return nil, err
	}

	return u.mapToResponse(project), nil
}

func (u *projectUseCase) DeleteProject(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *projectUseCase) mapToResponse(p *entity.Project) *model.ProjectResponse {
	return &model.ProjectResponse{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		UserID:         p.UserID,
		Name:           p.Name,
		Domain:         p.Domain,
		Status:         p.Status,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
