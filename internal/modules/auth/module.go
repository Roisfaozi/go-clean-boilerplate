package auth

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/repository"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/usecase"
	userRepository "github.com/Roisfaozi/casbin-db/internal/modules/user/repository"
	"github.com/Roisfaozi/casbin-db/internal/utils/jwt"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
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
) *AuthModule {
	tokenRepository := repository.NewTokenRepositoryRedis(redis, log)
	userRepo := userRepository.NewUserRepository(db, log)

	authUseCase := usecase.NewAuthUsecase(jwtManager, tokenRepository, userRepo, tm, log, wsManager)

	authHandler := http.NewAuthHandler(authUseCase, log, validator)

	return &AuthModule{
		handler: authHandler,
	}
}

func (m *AuthModule) AuthHandler() *http.AuthHandler {
	return m.handler
}
