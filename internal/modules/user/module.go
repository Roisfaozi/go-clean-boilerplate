package user

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/user/delivery/http"
	userRepository "github.com/Roisfaozi/casbin-db/internal/modules/user/repository"
	"github.com/Roisfaozi/casbin-db/internal/modules/user/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserModule struct {
	userHandler *http.UserHandler
}

// NewUserModule creates a new UserModule instance with the given dependencies.
//
// db: The GORM database connection.
// log: The logger instance.
// validator: The validator instance.
// tm: The transaction manager instance.
//
// Returns a pointer to the newly created UserModule instance.
func NewUserModule(db *gorm.DB, log *logrus.Logger, validator *validator.Validate, tm tx.WithTransactionManager) *UserModule {
	userRepository := userRepository.NewUserRepository(db, log)

	userUseCase := usecase.NewUserUseCase(log, validator, tm, userRepository)

	userHandler := http.NewUserHandler(userUseCase, log)

	return &UserModule{
		userHandler: userHandler,
	}
}

func (m *UserModule) UserHandler() *http.UserHandler {
	return m.userHandler
}
