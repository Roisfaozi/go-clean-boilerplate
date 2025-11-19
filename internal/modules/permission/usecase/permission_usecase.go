package usecase

import (
	"context"
	"errors"
	"github.com/Roisfaozi/casbin-db/internal/modules/role/repository"
	"github.com/Roisfaozi/casbin-db/internal/utils/exception"
	"github.com/casbin/casbin/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// IPermissionUseCase defines the interface for permission management.
type IPermissionUseCase interface {
	AssignRoleToUser(ctx context.Context, userID, role string) error
	GrantPermissionToRole(role, path, method string) error
	RevokePermissionFromRole(role, path, method string) error
	GetAllPermissions() ([][]string, error)
	GetPermissionsForRole(role string) ([][]string, error)
	UpdatePermission(oldPermission, newPermission []string) (bool, error)
}

// PermissionUseCase implements the permission use case.
type PermissionUseCase struct {
	enforcer *casbin.Enforcer
	log      *logrus.Logger
	RoleRepo repository.RoleRepository
}

// NewPermissionUseCase creates a new PermissionUseCase.
func NewPermissionUseCase(enforcer *casbin.Enforcer, log *logrus.Logger, roleRepo repository.RoleRepository) IPermissionUseCase {
	return &PermissionUseCase{
		enforcer: enforcer,
		log:      log,
		RoleRepo: roleRepo,
	}
}

// AssignRoleToUser assigns a role to a user after validating the role exists.
func (uc *PermissionUseCase) AssignRoleToUser(ctx context.Context, userID, role string) error {
	uc.log.Infof("Attempting to assign role '%s' to user '%s'", role, userID)

	// First, validate that the role exists in the roles table.
	_, err := uc.RoleRepo.FindByName(ctx, role)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.Warnf("Assign role failed: role '%s' does not exist.", role)
			return exception.ErrBadRequest // The provided role name is invalid.
		}
		uc.log.Errorf("Failed to query role repository: %v", err)
		return exception.ErrInternalServer
	}

	// If the role exists, proceed with assigning it in Casbin.
	uc.log.Infof("Role validated. Assigning role '%s' to user '%s'", role, userID)
	_, err = uc.enforcer.AddGroupingPolicy(userID, role)
	if err != nil {
		uc.log.Errorf("Failed to add grouping policy: %v", err)
		return err // Return the original error from casbin
	}
	return nil
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
	removed, err := uc.enforcer.RemovePolicy(role, path, method)
	if err != nil {
		return err
	}
	if !removed {
		return errors.New("policy to revoke not found")
	}
	return nil
}

// GetAllPermissions retrieves all policy rules from Casbin.
func (uc *PermissionUseCase) GetAllPermissions() ([][]string, error) {
	uc.log.Info("Retrieving all permissions")
	policies, err := uc.enforcer.GetPolicy()
	if err != nil {
		uc.log.Errorf("Failed to get all permissions: %v", err)
		return nil, err
	}
	return policies, nil
}

// GetPermissionsForRole retrieves all policy rules for a specific role.
func (uc *PermissionUseCase) GetPermissionsForRole(role string) ([][]string, error) {
	uc.log.Infof("Retrieving permissions for role '%s'", role)
	policies, err := uc.enforcer.GetFilteredPolicy(0, role)
	if err != nil {
		uc.log.Errorf("Failed get permission for role '%s'", role)

		return nil, err
	}
	return policies, nil
}

// UpdatePermission removes an old policy and adds a new one.
func (uc *PermissionUseCase) UpdatePermission(oldPermission, newPermission []string) (bool, error) {
	if len(oldPermission) == 0 || len(newPermission) == 0 {
		uc.log.Errorf("Old permission '%s' and new permission '%s'", oldPermission, newPermission)
		return false, errors.New("old and new permissions cannot be empty")
	}

	uc.log.Infof("Updating permission from %v to %v", oldPermission, newPermission)

	// Casbin's UpdatePolicy internally checks if the old policy exists.
	// It returns false if the old policy doesn't exist, which we can wrap in an error.
	updated, err := uc.enforcer.UpdatePolicy(oldPermission, newPermission)
	if err != nil {
		uc.log.Errorf("Failed update permission: %v", err)
		return false, err
	}
	if !updated {
		uc.log.Errorf("Policy to update not found: %v", oldPermission)
		return false, errors.New("policy to update not found")
	}

	return true, nil
}
