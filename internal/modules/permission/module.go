package permission

import (
	"github.com/Roisfaozi/casbin-db/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/permission/usecase"
	roleRepository "github.com/Roisfaozi/casbin-db/internal/modules/role/repository"
	"github.com/casbin/casbin/v2"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type PermissionModule struct {
	permissionHandler *http.PermissionHandler
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

	permissionHandler := http.NewPermissionHandler(permissionUseCase, validate, log)

	return &PermissionModule{
		permissionHandler: permissionHandler,
	}
}

func (m *PermissionModule) PermissionHandler() *http.PermissionHandler {
	return m.permissionHandler
}
