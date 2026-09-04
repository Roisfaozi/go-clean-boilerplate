package project

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type ProjectModule struct {
	ProjectController *http.ProjectController
	ProjectRepo       repository.ProjectRepository
}

func NewProjectModule(db *sqlx.DB, validate *validator.Validate) *ProjectModule {
	repo := repository.NewProjectRepository(db)
	uc := usecase.NewProjectUseCase(repo)
	ctrl := http.NewProjectController(uc, validate)

	return &ProjectModule{
		ProjectController: ctrl,
		ProjectRepo:       repo,
	}
}
