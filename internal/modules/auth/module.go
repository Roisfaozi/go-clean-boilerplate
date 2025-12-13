package auth

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/usecase"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase" // New Import
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
	handler *http.AuthHandler
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
) *AuthModule {
	tokenRepository := repository.NewTokenRepositoryRedis(redis, log)
	userRepo := userRepository.NewUserRepository(db, log)

	authUseCase := usecase.NewAuthUsecase(jwtManager, tokenRepository, userRepo, tm, log, wsManager, enforcer) // Pass enforcer

	authHandler := http.NewAuthHandler(authUseCase, log, validator)

	return &AuthModule{
		handler: authHandler,
	}
}

func (m *AuthModule) AuthHandler() *http.AuthHandler {
	return m.handler
}
