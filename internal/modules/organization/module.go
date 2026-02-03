package organization

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OrganizationModule encapsulates all organization-related dependencies
type OrganizationModule struct {
	OrganizationController *http.OrganizationController
	OrgRepo                repository.OrganizationRepository
	MemberRepo             repository.OrganizationMemberRepository
}

// NewOrganizationModule creates a new OrganizationModule with all dependencies wired
func NewOrganizationModule(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	tm tx.WithTransactionManager,
) *OrganizationModule {
	// Create repositories
	orgRepo := repository.NewOrganizationRepository(db)
	memberRepo := repository.NewOrganizationMemberRepository(db)

	// Create use cases
	orgUseCase := usecase.NewOrganizationUseCase(log, tm, orgRepo, memberRepo)
	memberUseCase := usecase.NewOrganizationMemberUseCase(log, tm, memberRepo, orgRepo)

	// Create controller
	orgController := http.NewOrganizationController(orgUseCase, memberUseCase, log, validate)

	return &OrganizationModule{
		OrganizationController: orgController,
		OrgRepo:                orgRepo,
		MemberRepo:             memberRepo,
	}
}

// Controller returns the organization controller
func (m *OrganizationModule) Controller() *http.OrganizationController {
	return m.OrganizationController
}
