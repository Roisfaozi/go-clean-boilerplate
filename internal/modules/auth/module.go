package auth

import (
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/casbin/casbin/v2"
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
	enforcer *casbin.Enforcer,
	auditModule *audit.AuditModule,
	taskDistributor worker.TaskDistributor,
	orgRepo orgRepo.OrganizationRepository, // Injected dependency
	ticketManager ws.TicketManager,
) *AuthModule {
	tokenRepo := repository.NewTokenRepositoryRedis(redisClient, log, db)
	userRepository := userRepo.NewUserRepository(db, log)

	authUseCase := usecase.NewAuthUsecase(maxLoginAttempts, lockoutDuration, jwtManager, tokenRepo, userRepository, orgRepo, tm, log, wsManager, sseManager, enforcer, auditModule.AuditController.UseCase, taskDistributor, ticketManager)
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
