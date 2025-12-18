package user

import (
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserModule struct {
	userController *http.UserController
}

// NewUserModule creates a new UserModule instance with the given dependencies.
//
// db: The GORM database connection.
// log: The logger instance.
// validator: The validator instance.
// tm: The transaction manager instance.
// enforcer: The Casbin enforcer instance.
//
// Returns a pointer to the newly created UserModule instance.
func NewUserModule(db *gorm.DB, log *logrus.Logger, validator *validator.Validate, tm tx.WithTransactionManager, enforcer permissionUseCase.IEnforcer) *UserModule {
	userRepository := userRepository.NewUserRepository(db, log)

	userUseCase := usecase.NewUserUseCase(log, tm, userRepository, enforcer)

	userController := http.NewUserController(userUseCase, log, validator)

	return &UserModule{
		userController: userController,
	}
}

func (m *UserModule) UserController() *http.UserController {
	return m.userController
}
