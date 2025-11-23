package access

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/repository"
	"github.com/Roisfaozi/casbin-db/internal/modules/access/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AccessModule struct {
	accessHandler *http.AccessHandler
}

func NewAccessModule(db *gorm.DB, log *logrus.Logger, validate *validator.Validate) *AccessModule {
	repo := repository.NewAccessRepository(db)
	uc := usecase.NewAccessUseCase(repo, log)
	handler := http.NewAccessHandler(uc, validate, log)

	return &AccessModule{
		accessHandler: handler,
	}
}

func (m *AccessModule) AccessHandler() *http.AccessHandler {
	return m.accessHandler
}
