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
	accessHandler *http.AccessHandler
}

// NewAccessModule creates a new AccessModule instance with the given dependencies.
//
// db: The GORM database connection.
// log: The logger instance.
// validate: The validator instance.
//
// Returns a pointer to the newly created AccessModule instance.
func NewAccessModule(db *gorm.DB, log *logrus.Logger, validate *validator.Validate) *AccessModule {
	repo := repository.NewAccessRepository(db, log)
	uc := usecase.NewAccessUseCase(repo, log)
	handler := http.NewAccessHandler(uc, validate, log)

	return &AccessModule{
		accessHandler: handler,
	}
}

func (m *AccessModule) AccessHandler() *http.AccessHandler {
	return m.accessHandler
}
