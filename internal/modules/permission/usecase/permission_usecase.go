package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IPermissionUseCase interface {
	AssignRoleToUser(ctx context.Context, userID, role string) error
	GrantPermissionToRole(ctx context.Context, role, path, method string) error
	RevokePermissionFromRole(ctx context.Context, role, path, method string) error
	GetAllPermissions(ctx context.Context) ([][]string, error)
	GetPermissionsForRole(ctx context.Context, role string) ([][]string, error)
	UpdatePermission(ctx context.Context, oldPermission, newPermission []string) (bool, error)
}

type PermissionUseCase struct {
	enforcer IEnforcer
	log      *logrus.Logger
	RoleRepo repository.RoleRepository
}

func NewPermissionUseCase(enforcer IEnforcer, log *logrus.Logger, roleRepo repository.RoleRepository) IPermissionUseCase {
	return &PermissionUseCase{
		enforcer: enforcer,
		log:      log,
		RoleRepo: roleRepo,
	}
}

func (uc *PermissionUseCase) AssignRoleToUser(ctx context.Context, userID, role string) error {
	uc.log.WithContext(ctx).Infof("Attempting to assign role '%s' to user '%s'", role, userID)

	if userID == "" || role == "" {
		return fmt.Errorf("userID and role are required")
	}

	_, err := uc.RoleRepo.FindByName(ctx, role)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("Assign role failed: role '%s' does not exist.", role)
			return exception.ErrBadRequest
		}
		uc.log.WithContext(ctx).Errorf("Failed to query role repository: %v", err)
		return exception.ErrInternalServer
	}

	uc.log.WithContext(ctx).Infof("Role validated. Removing existing roles and assigning role '%s' to user '%s'", role, userID)

	_, err = uc.enforcer.RemoveFilteredGroupingPolicy(0, userID)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to remove existing roles: %v", err)
		return exception.ErrInternalServer
	}

	_, err = uc.enforcer.AddGroupingPolicy(userID, role)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to add grouping policy: %v", err)
		return err
	}
	return nil
}

func (uc *PermissionUseCase) GrantPermissionToRole(ctx context.Context, role, path, method string) error {
	uc.log.WithContext(ctx).Infof("Attempting to grant permission to role '%s'", role)

	if role == "" || path == "" || method == "" {
		return fmt.Errorf("role, path, and method are required")
	}

	_, err := uc.RoleRepo.FindByName(ctx, role)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("Grant permission failed: role '%s' does not exist.", role)
			return exception.ErrBadRequest
		}
		uc.log.WithContext(ctx).Errorf("Failed to query role repository for GrantPermission: %v", err)
		return exception.ErrInternalServer
	}

	uc.log.WithContext(ctx).Infof("Granting permission to role '%s' for %s %s", role, method, path)
	_, err = uc.enforcer.AddPolicy(role, path, method)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to add policy: %v", err)
		return err
	}
	return nil
}

func (uc *PermissionUseCase) RevokePermissionFromRole(ctx context.Context, role, path, method string) error {
	uc.log.WithContext(ctx).Infof("Attempting to revoke permission from role '%s'", role)

	if role == "" || path == "" || method == "" {
		return fmt.Errorf("role, path, and method are required")
	}

	_, err := uc.RoleRepo.FindByName(ctx, role)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("Revoke permission failed: role '%s' does not exist.", role)
			return exception.ErrBadRequest
		}
		uc.log.WithContext(ctx).Errorf("Failed to query role repository for RevokePermission: %v", err)
		return exception.ErrInternalServer
	}

	uc.log.WithContext(ctx).Infof("Revoking permission from role '%s' for %s %s", role, method, path)
	removed, err := uc.enforcer.RemovePolicy(role, path, method)
	if err != nil {
		return err
	}
	if !removed {
		return errors.New("policy to revoke not found")
	}
	return nil
}

func (uc *PermissionUseCase) GetAllPermissions(ctx context.Context) ([][]string, error) {
	uc.log.WithContext(ctx).Info("Retrieving all permissions")
	policies, err := uc.enforcer.GetPolicy()
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get all permissions: %v", err)
		return nil, err
	}
	return policies, nil
}

func (uc *PermissionUseCase) GetPermissionsForRole(ctx context.Context, role string) ([][]string, error) {
	uc.log.WithContext(ctx).Infof("Retrieving permissions for role '%s'", role)
	policies, err := uc.enforcer.GetFilteredPolicy(0, role)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed get permission for role '%s'", role)
		return nil, err
	}
	return policies, nil
}

func (uc *PermissionUseCase) UpdatePermission(ctx context.Context, oldPermission, newPermission []string) (bool, error) {
	if len(oldPermission) == 0 || len(newPermission) == 0 {
		return false, errors.New("old and new permissions cannot be empty")
	}

	uc.log.WithContext(ctx).Infof("Updating permission from %v to %v", oldPermission, newPermission)
	updated, err := uc.enforcer.UpdatePolicy(oldPermission, newPermission)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed update permission: %v", err)
		return false, err
	}
	if !updated {
		uc.log.WithContext(ctx).Errorf("Policy to update not found: %v", oldPermission)
		return false, errors.New("policy to update not found")
	}

	return true, nil
}
