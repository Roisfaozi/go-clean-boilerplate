package audit

import (
	auditHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuditModule struct {
	AuditHandler *auditHttp.AuditHandler
	AuditUseCase usecase.AuditUseCase
	AuditRepo    usecase.AuditRepository
}

func NewAuditModule(db *gorm.DB, log *logrus.Logger) *AuditModule {
	repo := repository.NewAuditRepository(db, log)
	uc := usecase.NewAuditUseCase(repo, log)
	handler := auditHttp.NewAuditHandler(uc, log)

	return &AuditModule{
		AuditHandler: handler,
		AuditUseCase: uc,
		AuditRepo:    repo,
	}
}
