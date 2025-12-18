package permission

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/casbin/casbin/v2"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type PermissionModule struct {
	permissionController *http.PermissionController
}

// NewPermissionModule creates a new instance of PermissionModule.
//
// It takes the following parameters:
// - enforcer: the casbin.Enforcer implementation.
// - validate: the validator.Validate implementation.
// - log: the logrus.Logger implementation.
// - roleRepo: the roleRepository.RoleRepository implementation.
//
// It returns a pointer to the newly created PermissionModule.
func NewPermissionModule(enforcer *casbin.Enforcer, validate *validator.Validate, log *logrus.Logger, roleRepo roleRepository.RoleRepository) *PermissionModule {

	permissionUseCase := usecase.NewPermissionUseCase(enforcer, log, roleRepo)

	permissionController := http.NewPermissionController(permissionUseCase, validate, log)

	return &PermissionModule{
		permissionController: permissionController,
	}
}

func (m *PermissionModule) PermissionController() *http.PermissionController {
	return m.permissionController
}
