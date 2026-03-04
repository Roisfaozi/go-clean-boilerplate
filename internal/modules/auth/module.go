package auth

import (
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/util"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuthModule struct {
	AuthController *http.AuthController
	AuthUseCase    usecase.AuthUseCase
	TokenRepo      repository.TokenRepository
}

func NewAuthModule(
	maxLoginAttempts int,
	lockoutDuration time.Duration,
	jwtManager *jwt.JWTManager,
	db *gorm.DB,
	redisClient *redis.Client,
	log *logrus.Logger,
	validate *validator.Validate,
	tm tx.WithTransactionManager,
	wsManager ws.Manager,
	sseManager *sse.Manager,
	enforcer permissionUseCase.IEnforcer,
	auditModule *audit.AuditModule,
	taskDistributor worker.TaskDistributor,
	orgRepo orgRepo.OrganizationRepository,
	ticketManager ws.TicketManager,
	defaultRole string,
	defaultDomain string,
) *AuthModule {
	tokenRepo := repository.NewTokenRepositoryRedis(redisClient, log, db, &util.RealClock{})
	userRepository := userRepo.NewUserRepository(db, log)

	publisher := delivery.NewEventPublisher(wsManager, sseManager, log)
	authz := repository.NewCasbinAdapter(enforcer, defaultRole, defaultDomain)

	authUseCase := usecase.NewAuthUsecase(
		maxLoginAttempts,
		lockoutDuration,
		jwtManager,
		tokenRepo,
		userRepository,
		orgRepo,
		tm,
		log,
		publisher,
		authz,
		taskDistributor,
		ticketManager,
	)
	authController := http.NewAuthController(authUseCase, log, validate)

	return &AuthModule{
		AuthController: authController,
		AuthUseCase:    authUseCase,
		TokenRepo:      tokenRepo,
	}
}

func (m *AuthModule) Controller() *http.AuthController {
	return m.AuthController
}
