package role

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoleModule struct {
	RoleController *http.RoleController
}

func NewRoleModule(db *gorm.DB, log *logrus.Logger, validator *validator.Validate, tm tx.WithTransactionManager) *RoleModule {
	roleRepo := repository.NewRoleRepository(db, log)
	roleUseCase := usecase.NewRoleUseCase(log, tm, roleRepo)
	roleController := http.NewRoleController(roleUseCase, log, validator)

	return &RoleModule{
		RoleController: roleController,
	}
}

func (m *RoleModule) Controller() *http.RoleController {
	return m.RoleController
}
