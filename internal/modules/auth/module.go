package auth

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuthModule struct {
	AuthController *http.AuthController
}

func NewAuthModule(
	jwtManager *jwt.JWTManager,
	db *gorm.DB,
	redis *redis.Client,
	log *logrus.Logger,
	validator *validator.Validate,
	tm tx.WithTransactionManager,
	wsManager ws.Manager,
	enforcer permissionUseCase.IEnforcer,
	auditModule *audit.AuditModule,
) *AuthModule {
	tokenRepository := repository.NewTokenRepositoryRedis(redis, log)
	userRepo := userRepository.NewUserRepository(db, log)

	authUseCase := usecase.NewAuthUsecase(
		jwtManager, 
		tokenRepository, 
		userRepo, 
		tm, 
		log, 
		wsManager, 
		enforcer, 
		auditModule.AuditUseCase,
	)

	authController := http.NewAuthController(authUseCase, log, validator)

	return &AuthModule{
		AuthController: authController,
	}
}

func (m *AuthModule) Controller() *http.AuthController {
	return m.AuthController
}