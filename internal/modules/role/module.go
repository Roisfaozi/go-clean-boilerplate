package role

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	roleRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoleModule struct {
	roleController *http.RoleController
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
	roleUseCase := usecase.NewRoleUseCase(log, tm, roleRepo)
	roleController := http.NewRoleController(roleUseCase, log, validator)

	return &RoleModule{
		roleController: roleController,
	}
}

func (m *RoleModule) RoleController() *http.RoleController {
	return m.roleController
}
