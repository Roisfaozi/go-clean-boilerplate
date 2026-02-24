package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	roleRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IPermissionUseCase interface {
	AssignRoleToUser(ctx context.Context, userID, role string) error
	RevokeRoleFromUser(ctx context.Context, userID, role string) error
	GrantPermissionToRole(ctx context.Context, role, path, method string) error
	RevokePermissionFromRole(ctx context.Context, role, path, method string) error
	GetAllPermissions(ctx context.Context) ([][]string, error)
	GetPermissionsForRole(ctx context.Context, role string) ([][]string, error)
	UpdatePermission(ctx context.Context, oldPermission, newPermission []string) (bool, error)
	GetUsersForRole(ctx context.Context, role string) ([]string, error)

	AddParentRole(ctx context.Context, childRole, parentRole string) error
	RemoveParentRole(ctx context.Context, childRole, parentRole string) error
	GetParentRoles(ctx context.Context, role string) ([]string, error)

	BatchCheckPermission(ctx context.Context, userID string, items []model.PermissionCheckItem) (map[string]bool, error)

	// New methods for Matrix View
	GetResourceAggregation(ctx context.Context) (*model.ResourceAggregationResponse, error)
	GetInheritanceTree(ctx context.Context) (*model.InheritanceTreeResponse, error)
}

type PermissionUseCase struct {
	enforcer IEnforcer
	log      *logrus.Logger
	RoleRepo roleRepository.RoleRepository
	UserRepo userRepository.UserRepository
}

func NewPermissionUseCase(enforcer IEnforcer, log *logrus.Logger, roleRepo roleRepository.RoleRepository, userRepo userRepository.UserRepository) IPermissionUseCase {
	return &PermissionUseCase{
		enforcer: enforcer,
		log:      log,
		RoleRepo: roleRepo,
		UserRepo: userRepo,
	}
}

func (uc *PermissionUseCase) BatchCheckPermission(ctx context.Context, userID string, items []model.PermissionCheckItem) (map[string]bool, error) {
	results := make(map[string]bool)
	enf := uc.enforcer.WithContext(ctx)

	for _, item := range items {
		key := fmt.Sprintf("%s:%s", item.Resource, item.Action)

		// For now, batch check defaults to global domain.
		// Future: support domain in PermissionCheckItem
		allowed, err := enf.Enforce(userID, "global", item.Resource, item.Action)
		if err != nil {
			uc.log.WithContext(ctx).Errorf("Enforce error for %s on %s in domain global: %v", userID, item.Resource, err)
			results[key] = false
			continue
		}
		results[key] = allowed
	}

	return results, nil
}

func (uc *PermissionUseCase) AddParentRole(ctx context.Context, childRole, parentRole string) error {
	uc.log.WithContext(ctx).Infof("Adding inheritance: role '%s' inherits from '%s'", childRole, parentRole)

	if _, err := uc.RoleRepo.FindByName(ctx, childRole); err != nil {
		return exception.ErrBadRequest
	}
	if _, err := uc.RoleRepo.FindByName(ctx, parentRole); err != nil {
		return exception.ErrBadRequest
	}

	if childRole == parentRole {
		return errors.New("role cannot inherit from itself")
	}

	// Use 'global' domain for role inheritance by default
	_, err := uc.enforcer.WithContext(ctx).AddGroupingPolicy(childRole, parentRole, "global")
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to add parent role: %v", err)
		return err
	}
	return nil
}

func (uc *PermissionUseCase) RemoveParentRole(ctx context.Context, childRole, parentRole string) error {
	uc.log.WithContext(ctx).Infof("Removing inheritance: role '%s' inherits from '%s'", childRole, parentRole)

	// Filter by child, parent, and 'global' domain (v2)
	removed, err := uc.enforcer.WithContext(ctx).RemoveFilteredGroupingPolicy(0, childRole, parentRole, "global")
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to remove parent role: %v", err)
		return err
	}
	if !removed {
		return errors.New("inheritance relationship not found")
	}
	return nil
}

func (uc *PermissionUseCase) GetParentRoles(ctx context.Context, role string) ([]string, error) {
	// Defaults to 'global' domain
	roles, err := uc.enforcer.WithContext(ctx).GetRolesForUser(role, "global")
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get parent roles: %v", err)
		return nil, err
	}
	return roles, nil
}

func (uc *PermissionUseCase) AssignRoleToUser(ctx context.Context, userID, role string) error {
	uc.log.WithContext(ctx).Infof("Attempting to assign role '%s' to user '%s'", role, userID)

	if userID == "" || role == "" {
		return fmt.Errorf("userID and role are required")
	}

	_, err := uc.UserRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("Assign role failed: user '%s' does not exist.", userID)
			return exception.ErrNotFound
		}
		uc.log.WithContext(ctx).Errorf("Failed to query user repository: %v", err)
		return exception.ErrInternalServer
	}

	_, err = uc.RoleRepo.FindByName(ctx, role)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("Assign role failed: role '%s' does not exist.", role)
			return exception.ErrBadRequest
		}
		uc.log.WithContext(ctx).Errorf("Failed to query role repository: %v", err)
		return exception.ErrInternalServer
	}

	uc.log.WithContext(ctx).Infof("User and Role validated. Removing existing roles and assigning role '%s' to user '%s' in domain 'global'", role, userID)

	enf := uc.enforcer.WithContext(ctx)

	// Remove all roles for this user in ALL domains? Or just global?
	// For legacy compatibility, let's just clear global roles.
	_, err = enf.RemoveFilteredGroupingPolicy(0, userID, "", "global")
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to remove existing global roles: %v", err)
		return exception.ErrInternalServer
	}

	_, err = enf.AddGroupingPolicy(userID, role, "global")
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to add grouping policy: %v", err)
		return err
	}
	return nil
}

func (uc *PermissionUseCase) RevokeRoleFromUser(ctx context.Context, userID, role string) error {
	uc.log.WithContext(ctx).Infof("Attempting to revoke role '%s' from user '%s'", role, userID)

	if userID == "" || role == "" {
		return fmt.Errorf("userID and role are required")
	}

	_, err := uc.UserRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("Revoke role failed: user '%s' does not exist.", userID)
			return exception.ErrNotFound
		}
		uc.log.WithContext(ctx).Errorf("Failed to query user repository: %v", err)
		return exception.ErrInternalServer
	}

	_, err = uc.RoleRepo.FindByName(ctx, role)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.WithContext(ctx).Warnf("Revoke role failed: role '%s' does not exist.", role)
			return exception.ErrBadRequest
		}
		uc.log.WithContext(ctx).Errorf("Failed to query role repository: %v", err)
		return exception.ErrInternalServer
	}

	// Filter by user, role, and 'global' domain
	removed, err := uc.enforcer.WithContext(ctx).RemoveFilteredGroupingPolicy(0, userID, role, "global")
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to remove role from user: %v", err)
		return exception.ErrInternalServer
	}
	if !removed {
		return errors.New("role was not assigned to user in domain global")
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

	uc.log.WithContext(ctx).Infof("Granting permission to role '%s' for %s %s in domain 'global'", role, method, path)
	_, err = uc.enforcer.WithContext(ctx).AddPolicy(role, "global", path, method)
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

	uc.log.WithContext(ctx).Infof("Revoking permission from role '%s' for %s %s in domain 'global'", role, method, path)
	removed, err := uc.enforcer.WithContext(ctx).RemovePolicy(role, "global", path, method)
	if err != nil {
		return err
	}
	if !removed {
		return errors.New("policy to revoke not found in domain global")
	}
	return nil
}

func (uc *PermissionUseCase) GetAllPermissions(ctx context.Context) ([][]string, error) {
	uc.log.WithContext(ctx).Info("Retrieving all permissions")
	policies, err := uc.enforcer.WithContext(ctx).GetPolicy()
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get all permissions: %v", err)
		return nil, err
	}
	return policies, nil
}

func (uc *PermissionUseCase) GetPermissionsForRole(ctx context.Context, role string) ([][]string, error) {
	uc.log.WithContext(ctx).Infof("Retrieving permissions for role '%s'", role)
	policies, err := uc.enforcer.WithContext(ctx).GetFilteredPolicy(0, role)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed get permission for role '%s'", role)
		return nil, err
	}
	return policies, nil
}

func (uc *PermissionUseCase) GetUsersForRole(ctx context.Context, role string) ([]string, error) {
	uc.log.WithContext(ctx).Infof("Retrieving users for role '%s'", role)
	// Casbin GetUsersForRole returns users assigned to this role in domain 'global'
	users, err := uc.enforcer.WithContext(ctx).GetUsersForRole(role, "global")
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get users for role '%s': %v", role, err)
		return nil, err
	}
	return users, nil
}

func (uc *PermissionUseCase) UpdatePermission(ctx context.Context, oldPermission, newPermission []string) (bool, error) {
	if len(oldPermission) == 0 || len(newPermission) == 0 {
		return false, errors.New("old and new permissions cannot be empty")
	}

	uc.log.WithContext(ctx).Infof("Updating permission from %v to %v", oldPermission, newPermission)
	updated, err := uc.enforcer.WithContext(ctx).UpdatePolicy(oldPermission, newPermission)
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
