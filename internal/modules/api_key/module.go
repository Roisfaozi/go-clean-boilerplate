package api_key

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ApiKeyModule struct {
	Repo       repository.ApiKeyRepository
	UseCase    usecase.ApiKeyUseCase
	Controller *http.ApiKeyController
}

func NewApiKeyModule(db *gorm.DB, log *logrus.Logger, validator *validator.Validate) *ApiKeyModule {
	repo := repository.NewApiKeyRepository(db)
	useCase := usecase.NewApiKeyUseCase(repo, log)
	controller := http.NewApiKeyController(useCase, log, validator)

	return &ApiKeyModule{
		Repo:       repo,
		UseCase:    useCase,
		Controller: controller,
	}
}
