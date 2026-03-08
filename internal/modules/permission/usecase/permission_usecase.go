package usecase

import (
	"context"
	"errors"
	"fmt"

	accessRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	auditUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	roleRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IPermissionUseCase interface {
	AssignRoleToUser(ctx context.Context, userID, role, domain string) error
	RevokeRoleFromUser(ctx context.Context, userID, role, domain string) error
	GrantPermissionToRole(ctx context.Context, role, path, method, domain string) error
	RevokePermissionFromRole(ctx context.Context, role, path, method, domain string) error
	GetAllPermissions(ctx context.Context) ([][]string, error)
	GetPermissionsForRole(ctx context.Context, role string) ([][]string, error)
	UpdatePermission(ctx context.Context, oldPermission, newPermission []string) (bool, error)
	GetUsersForRole(ctx context.Context, role, domain string) ([]string, error)

	AddParentRole(ctx context.Context, childRole, parentRole, domain string) error
	RemoveParentRole(ctx context.Context, childRole, parentRole, domain string) error
	GetParentRoles(ctx context.Context, role, domain string) ([]string, error)

	BatchCheckPermission(ctx context.Context, userID string, items []model.PermissionCheckItem) (map[string]bool, error)

	// Matrix View
	GetResourceAggregation(ctx context.Context) (*model.ResourceAggregationResponse, error)
	GetInheritanceTree(ctx context.Context) (*model.InheritanceTreeResponse, error)

	// Bulk Access Right assignment
	GetRoleAccessRights(ctx context.Context, role, domain string) ([]model.RoleAccessRightStatus, error)
	AssignAccessRight(ctx context.Context, req model.AssignAccessRightRequest) error
	RevokeAccessRight(ctx context.Context, req model.AssignAccessRightRequest) error
	DeleteRole(ctx context.Context, roleName string) error
}

type PermissionUseCase struct {
	enforcer   IEnforcer
	log        *logrus.Logger
	RoleRepo   roleRepository.RoleRepository
	UserRepo   userRepository.UserRepository
	AccessRepo accessRepository.AccessRepository
	AuditUC    auditUseCase.AuditUseCase
}

func NewPermissionUseCase(
	enforcer IEnforcer,
	log *logrus.Logger,
	roleRepo roleRepository.RoleRepository,
	userRepo userRepository.UserRepository,
	accessRepo accessRepository.AccessRepository,
	auditUC auditUseCase.AuditUseCase,
) IPermissionUseCase {
	return &PermissionUseCase{
		enforcer:   enforcer,
		log:        log,
		RoleRepo:   roleRepo,
		UserRepo:   userRepo,
		AccessRepo: accessRepo,
		AuditUC:    auditUC,
	}
}

func (uc *PermissionUseCase) BatchCheckPermission(ctx context.Context, userID string, items []model.PermissionCheckItem) (map[string]bool, error) {
	results := make(map[string]bool)
	enf := uc.enforcer.WithContext(ctx)

	for _, item := range items {
		key := fmt.Sprintf("%s:%s", item.Resource, item.Action)

		domain := item.Domain
		if domain == "" {
			domain = "global"
		}

		allowed, err := enf.Enforce(userID, domain, item.Resource, item.Action)
		if err != nil {
			uc.log.WithContext(ctx).Errorf("Enforce error for %s on %s in domain %s: %v", userID, item.Resource, domain, err)
			results[key] = false
			continue
		}
		results[key] = allowed
	}

	return results, nil
}

func (uc *PermissionUseCase) AddParentRole(ctx context.Context, childRole, parentRole, domain string) error {
	if domain == "" {
		domain = "global"
	}
	uc.log.WithContext(ctx).Infof("Adding inheritance: role '%s' inherits from '%s' in domain '%s'", childRole, parentRole, domain)

	if _, err := uc.RoleRepo.FindByName(ctx, childRole); err != nil {
		return exception.ErrBadRequest
	}
	if _, err := uc.RoleRepo.FindByName(ctx, parentRole); err != nil {
		return exception.ErrBadRequest
	}

	if childRole == parentRole {
		return errors.New("role cannot inherit from itself")
	}

	_, err := uc.enforcer.WithContext(ctx).AddGroupingPolicy(childRole, parentRole, domain)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to add parent role: %v", err)
		return err
	}
	return nil
}

func (uc *PermissionUseCase) RemoveParentRole(ctx context.Context, childRole, parentRole, domain string) error {
	if domain == "" {
		domain = "global"
	}
	uc.log.WithContext(ctx).Infof("Removing inheritance: role '%s' inherits from '%s' in domain '%s'", childRole, parentRole, domain)

	removed, err := uc.enforcer.WithContext(ctx).RemoveFilteredGroupingPolicy(0, childRole, parentRole, domain)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to remove parent role: %v", err)
		return err
	}
	if !removed {
		return errors.New("inheritance relationship not found")
	}
	return nil
}

func (uc *PermissionUseCase) GetParentRoles(ctx context.Context, role, domain string) ([]string, error) {
	if domain == "" {
		domain = "global"
	}
	roles, err := uc.enforcer.WithContext(ctx).GetRolesForUser(role, domain)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get parent roles: %v", err)
		return nil, err
	}
	return roles, nil
}

func (uc *PermissionUseCase) AssignRoleToUser(ctx context.Context, userID, role, domain string) error {
	if domain == "" {
		domain = "global"
	}
	uc.log.WithContext(ctx).Infof("Attempting to assign role '%s' to user '%s' in domain '%s'", role, userID, domain)

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

	uc.log.WithContext(ctx).Infof("User and Role validated. Removing existing roles and assigning role '%s' to user '%s' in domain '%s'", role, userID, domain)

	enf := uc.enforcer.WithContext(ctx)

	// Remove existing roles for this user in the specified domain
	_, err = enf.RemoveFilteredGroupingPolicy(0, userID, "", domain)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to remove existing roles in domain '%s': %v", domain, err)
		return exception.ErrInternalServer
	}

	_, err = enf.AddGroupingPolicy(userID, role, domain)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to add grouping policy: %v", err)
		return err
	}
	return nil
}

func (uc *PermissionUseCase) RevokeRoleFromUser(ctx context.Context, userID, role, domain string) error {
	if domain == "" {
		domain = "global"
	}
	uc.log.WithContext(ctx).Infof("Attempting to revoke role '%s' from user '%s' in domain '%s'", role, userID, domain)

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

	removed, err := uc.enforcer.WithContext(ctx).RemoveFilteredGroupingPolicy(0, userID, role, domain)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to remove role from user: %v", err)
		return exception.ErrInternalServer
	}
	if !removed {
		return errors.New("role was not assigned to user in specified domain")
	}
	return nil
}

func (uc *PermissionUseCase) GrantPermissionToRole(ctx context.Context, role, path, method, domain string) error {
	if domain == "" {
		domain = "global"
	}
	uc.log.WithContext(ctx).Infof("Attempting to grant permission to role '%s' in domain '%s'", role, domain)

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

	uc.log.WithContext(ctx).Infof("Granting permission to role '%s' for %s %s in domain '%s'", role, method, path, domain)
	_, err = uc.enforcer.WithContext(ctx).AddPolicy(role, domain, path, method)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to add policy: %v", err)
		return err
	}
	return nil
}

func (uc *PermissionUseCase) RevokePermissionFromRole(ctx context.Context, role, path, method, domain string) error {
	if domain == "" {
		domain = "global"
	}
	uc.log.WithContext(ctx).Infof("Attempting to revoke permission from role '%s' in domain '%s'", role, domain)

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

	uc.log.WithContext(ctx).Infof("Revoking permission from role '%s' for %s %s in domain '%s'", role, method, path, domain)
	removed, err := uc.enforcer.WithContext(ctx).RemovePolicy(role, domain, path, method)
	if err != nil {
		return err
	}
	if !removed {
		return errors.New("policy to revoke not found in specified domain")
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

func (uc *PermissionUseCase) GetUsersForRole(ctx context.Context, role, domain string) ([]string, error) {
	if domain == "" {
		domain = "global"
	}
	uc.log.WithContext(ctx).Infof("Retrieving users for role '%s' in domain '%s'", role, domain)
	users, err := uc.enforcer.WithContext(ctx).GetUsersForRole(role, domain)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get users for role '%s' in domain '%s': %v", role, domain, err)
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

func (uc *PermissionUseCase) DeleteRole(ctx context.Context, roleName string) error {
	uc.log.WithContext(ctx).Infof("Cleaning up Casbin policies for role '%s'", roleName)
	enf := uc.enforcer.WithContext(ctx)
	_, err := enf.DeleteRole(roleName)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to delete role from Casbin: %v", err)
		return err
	}

	// Force reload to ensure all enforcer instances (especially the global one used by middleware)
	// reflect the changes made in DB.
	if err := uc.enforcer.LoadPolicy(); err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to reload Casbin policy after role deletion: %v", err)
		// We don't return error here because the DB change was successful
	}

	return nil
}
