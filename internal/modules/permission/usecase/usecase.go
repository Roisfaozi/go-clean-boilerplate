package usecase

import (
	"github.com/casbin/casbin/v2"
	"github.com/sirupsen/logrus"
)

// IPermissionUseCase defines the interface for permission management.
type IPermissionUseCase interface {
	AssignRoleToUser(userID, role string) error
	GrantPermissionToRole(role, path, method string) error
	RevokePermissionFromRole(role, path, method string) error
}

// PermissionUseCase implements the permission use case.
type PermissionUseCase struct {
	enforcer *casbin.Enforcer
	log      *logrus.Logger
}

// NewPermissionUseCase creates a new PermissionUseCase.
func NewPermissionUseCase(enforcer *casbin.Enforcer, log *logrus.Logger) IPermissionUseCase {
	return &PermissionUseCase{
		enforcer: enforcer,
		log:      log,
	}
}

// AssignRoleToUser assigns a role to a user.
func (uc *PermissionUseCase) AssignRoleToUser(userID, role string) error {
	uc.log.Infof("Assigning role '%s' to user '%s'", role, userID)
	_, err := uc.enforcer.AddGroupingPolicy(userID, role)
	return err
}

// GrantPermissionToRole grants a permission to a role.
func (uc *PermissionUseCase) GrantPermissionToRole(role, path, method string) error {
	uc.log.Infof("Granting permission to role '%s' for %s %s", role, method, path)
	_, err := uc.enforcer.AddPolicy(role, path, method)
	return err
}

// RevokePermissionFromRole revokes a permission from a role.
func (uc *PermissionUseCase) RevokePermissionFromRole(role, path, method string) error {
	uc.log.Infof("Revoking permission from role '%s' for %s %s", role, method, path)
	_, err := uc.enforcer.RemovePolicy(role, path, method)
	return err
}
