package audit

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuditModule struct {
	AuditController *http.AuditController
	AuditUseCase    usecase.AuditUseCase
	AuditRepo       usecase.AuditRepository
}

// NewAuditModule creates a new instance of AuditModule.
//
// db: The GORM database connection.
// log: The logger instance.
// validate: The validator instance.
// wsManager: The WebSocket manager instance.
//
// Returns a pointer to the newly created AuditModule instance.
func NewAuditModule(db *gorm.DB, log *logrus.Logger, validate *validator.Validate, wsManager ws.Manager) *AuditModule {
	repo := repository.NewAuditRepository(db, log)
	uc := usecase.NewAuditUseCase(repo, log, wsManager)
	controller := http.NewAuditController(uc, validate, log)

	return &AuditModule{
		AuditController: controller,
		AuditUseCase:    uc,
		AuditRepo:       repo,
	}
}

func (m *AuditModule) Controller() *http.AuditController {
	return m.AuditController
}
