package organization

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OrganizationModule encapsulates all organization-related dependencies
type OrganizationModule struct {
	OrganizationController *http.OrganizationController
	OrgRepo                repository.OrganizationRepository
	MemberRepo             repository.OrganizationMemberRepository
	OrgReader              usecase.IOrganizationReader
	InvitationRepo         repository.InvitationRepository
	UserRepo               userRepo.UserRepository
}

// NewOrganizationModule creates a new OrganizationModule with all dependencies wired
func NewOrganizationModule(
	db *gorm.DB,
	redisClient *redis.Client,
	taskDistributor worker.TaskDistributor,
	userRepo userRepo.UserRepository,
	log *logrus.Logger,
	validate *validator.Validate,
	tm tx.WithTransactionManager,
	enforcer permissionUseCase.IEnforcer,
) *OrganizationModule {
	// Create repositories
	orgRepo := repository.NewOrganizationRepository(db)
	memberRepo := repository.NewOrganizationMemberRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)

	// Create cached organization reader
	orgReader := usecase.NewCachedOrgReader(memberRepo, redisClient, log)

	// Create use cases
	orgUseCase := usecase.NewOrganizationUseCase(log, tm, orgRepo, memberRepo, enforcer)
	memberUseCase := usecase.NewOrganizationMemberUseCase(log, tm, memberRepo, orgRepo, invitationRepo, userRepo, taskDistributor, enforcer)

	// Create controller
	orgController := http.NewOrganizationController(orgUseCase, memberUseCase, log, validate)

	return &OrganizationModule{
		OrganizationController: orgController,
		OrgRepo:                orgRepo,
		MemberRepo:             memberRepo,
		OrgReader:              orgReader,
		InvitationRepo:         invitationRepo,
		UserRepo:               userRepo,
	}
}

// Controller returns the organization controller
func (m *OrganizationModule) Controller() *http.OrganizationController {
	return m.OrganizationController
}

// Reader returns the cached organization reader for TenantMiddleware
func (m *OrganizationModule) Reader() usecase.IOrganizationReader {
	return m.OrgReader
}
