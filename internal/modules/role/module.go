package role

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/role/delivery/http"
	roleRepository "github.com/Roisfaozi/casbin-db/internal/modules/role/repository"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/usecase"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoleModule struct {
	Handler *http.RoleHandler
	Repo    roleRepository.RoleRepository
}

// NewRoleModule creates a new RoleModule instance with the given dependencies.
//
// db: The GORM database connection.
// log: The logger instance.
// validator: The validator instance.
// tm: The transaction manager instance.
//
// Returns a pointer to the newly created RoleModule instance.
func NewRoleModule(db *gorm.DB, log *logrus.Logger, validator *validator.Validate, tm tx.WithTransactionManager) *RoleModule {
	roleRepo := roleRepository.NewRoleRepository(db, log)
	roleUseCase := usecase.NewRoleUseCase(log, validator, tm, roleRepo)
	roleHandler := http.NewRoleHandler(roleUseCase, log)

	return &RoleModule{
		Handler: roleHandler,
		Repo:    roleRepo,
	}
}
