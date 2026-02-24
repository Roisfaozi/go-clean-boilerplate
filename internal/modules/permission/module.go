package permission

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type IEnforcer = usecase.IEnforcer

type PermissionModule struct {
	PermissionController *http.PermissionController
}

func NewPermissionModule(enforcer usecase.IEnforcer, validate *validator.Validate, log *logrus.Logger, roleRepo roleRepository.RoleRepository, userRepo userRepository.UserRepository) *PermissionModule {

	permissionUseCase := usecase.NewPermissionUseCase(enforcer, log, roleRepo, userRepo)

	permissionController := http.NewPermissionController(permissionUseCase, log, validate)

	return &PermissionModule{
		PermissionController: permissionController,
	}
}

func (m *PermissionModule) Controller() *http.PermissionController {
	return m.PermissionController
}
