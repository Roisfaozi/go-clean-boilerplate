package user

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/storage"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserModule struct {
	UserController *http.UserController
	UserRepo       userRepository.UserRepository
	UserUseCase    usecase.UserUseCase
}

func NewUserModule(
	db *gorm.DB,
	log *logrus.Logger,
	validator *validator.Validate,
	tm tx.WithTransactionManager,
	enforcer permissionUseCase.IEnforcer,
	auditModule *audit.AuditModule,
	authModule *auth.AuthModule,
	webhookModule *webhook.WebhookModule,
	storage storage.Provider,
) *UserModule {
	userRepo := userRepository.NewUserRepository(db, log)

	userUseCase := usecase.NewUserUseCase(tm, log, userRepo, enforcer, auditModule.AuditUseCase, authModule.AuthUseCase, webhookModule.UseCase, storage)

	userController := http.NewUserController(userUseCase, log, validator)

	return &UserModule{
		UserController: userController,
		UserRepo:       userRepo,
		UserUseCase:    userUseCase,
	}
}

func (m *UserModule) Controller() *http.UserController {
	return m.UserController
}
