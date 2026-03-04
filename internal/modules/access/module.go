package access

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AccessModule struct {
	AccessController *http.AccessController
	AccessRepo       repository.AccessRepository
}

func NewAccessModule(db *gorm.DB, log *logrus.Logger, validate *validator.Validate) *AccessModule {
	repo := repository.NewAccessRepository(db, log)
	uc := usecase.NewAccessUseCase(repo, log)
	controller := http.NewAccessController(uc, validate, log)

	return &AccessModule{
		AccessController: controller,
		AccessRepo:       repo,
	}
}

func (m *AccessModule) Controller() *http.AccessController {
	return m.AccessController
}
