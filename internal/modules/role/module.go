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

// RoleModule contains all the components of the role module.
type RoleModule struct {
	Handler *http.RoleHandler
	Repo    roleRepository.RoleRepository
}

// NewRoleModule initializes a new role module with all its dependencies.
func NewRoleModule(db *gorm.DB, log *logrus.Logger, validator *validator.Validate, tm tx.WithTransactionManager) *RoleModule {
	roleRepo := roleRepository.NewRoleRepository(db, log)
	roleUseCase := usecase.NewRoleUseCase(log, validator, tm, roleRepo)
	roleHandler := http.NewRoleHandler(roleUseCase, log)

	return &RoleModule{
		Handler: roleHandler,
		Repo:    roleRepo,
	}
}