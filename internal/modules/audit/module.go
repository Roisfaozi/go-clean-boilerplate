package audit

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuditModule struct {
	auditController *http.AuditController
}

// NewAuditModule creates a new instance of AuditModule.
//
// db: The GORM database connection.
// log: The logger instance.
//
// Returns a pointer to the newly created AuditModule instance.
func NewAuditModule(db *gorm.DB, log *logrus.Logger) *AuditModule {
	repo := repository.NewAuditRepository(db, log)
	uc := usecase.NewAuditUseCase(repo, log)
	controller := http.NewAuditController(uc, log)

	return &AuditModule{
		auditController: controller,
	}
}

func (m *AuditModule) AuditController() *http.AuditController {
	return m.auditController
}
