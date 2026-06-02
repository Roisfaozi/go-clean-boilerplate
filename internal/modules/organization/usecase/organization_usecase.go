package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model/converter"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultOwnerRoleID is the default role assigned to organization owners
	DefaultOwnerRoleID = "role:org-owner"
	adminRoleID        = "role:admin"
	defaultUserRoleID  = "role:user"
	globalDomain       = "global"
	superAdminRoleID   = "role:superadmin"
)

type organizationUseCase struct {
	Log        *logrus.Logger
	TM         tx.WithTransactionManager
	OrgRepo    repository.OrganizationRepository
	MemberRepo repository.OrganizationMemberRepository
	OrgReader  IOrganizationReader
	Enforcer   permissionUseCase.IEnforcer
}

// NewOrganizationUseCase creates a new OrganizationUseCase instance
func NewOrganizationUseCase(
	log *logrus.Logger,
	tm tx.WithTransactionManager,
	orgRepo repository.OrganizationRepository,
	memberRepo repository.OrganizationMemberRepository,
	orgReader IOrganizationReader,
	enforcer permissionUseCase.IEnforcer,
) OrganizationUseCase {
	return &organizationUseCase{
		Log:        log,
		TM:         tm,
		OrgRepo:    orgRepo,
		MemberRepo: memberRepo,
		OrgReader:  orgReader,
		Enforcer:   enforcer,
	}
}

// CreateOrganization creates a new organization with the current user as owner
func (uc *organizationUseCase) CreateOrganization(ctx context.Context, userID string, request *model.CreateOrganizationRequest) (*model.OrganizationResponse, error) {
	var response *model.OrganizationResponse

	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Check if slug is already taken
		exists, err := uc.OrgRepo.SlugExists(txCtx, request.Slug)
		if err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to check slug existence: %v", err)
			return exception.ErrInternalServer
		}
		if exists {
			uc.Log.WithContext(txCtx).Warnf("Slug %s already exists", request.Slug)
			return exception.ErrConflict
		}

		// Generate new ID
		newID, err := uuid.NewV7()
		if err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to generate UUID: %v", err)
			return exception.ErrInternalServer
		}

		// Create organization
		org := &entity.Organization{
			ID:      newID.String(),
			Name:    request.Name,
			Slug:    request.Slug,
			OwnerID: userID,
			Status:  entity.OrgStatusActive,
		}

		// Atomic create (org + owner member)
		if err := uc.OrgRepo.Create(txCtx, org, DefaultOwnerRoleID); err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to create organization: %v", err)
			return exception.ErrInternalServer
		}

		// Add Casbin Grouping Policy for owner in this org domain
		if uc.Enforcer != nil {
			enf := uc.Enforcer.WithContext(txCtx)
			if err := uc.bootstrapOrganizationPolicies(enf, org.ID); err != nil {
				uc.Log.WithContext(txCtx).Errorf("Failed to bootstrap organization policies: %v", err)
				return exception.ErrInternalServer
			}

			if _, err := enf.AddGroupingPolicy(userID, DefaultOwnerRoleID, org.ID); err != nil {
				uc.Log.WithContext(txCtx).Errorf("Failed to add Casbin grouping policy: %v", err)
				return exception.ErrInternalServer
			}
		}

		response = converter.OrganizationToResponse(org)
		return nil
	})
	if err == nil && uc.Enforcer != nil {
		if loadErr := uc.Enforcer.LoadPolicy(); loadErr != nil {
			uc.Log.WithContext(ctx).Errorf("Failed to reload Casbin policy after organization creation: %v", loadErr)
			return nil, exception.ErrInternalServer
		}
	}

	return response, err
}

func (uc *organizationUseCase) bootstrapOrganizationPolicies(enf permissionUseCase.IEnforcer, orgID string) error {
	defaultRoles := []string{adminRoleID, defaultUserRoleID}

	for _, roleID := range defaultRoles {
		policies, err := enf.GetFilteredPolicy(0, roleID, globalDomain)
		if err != nil {
			return err
		}

		for _, policy := range policies {
			if len(policy) < 4 {
				continue
			}

			if _, err := enf.AddPolicy(policy[0], orgID, policy[2], policy[3]); err != nil {
				return err
			}
		}
	}

	if _, err := enf.AddGroupingPolicy(DefaultOwnerRoleID, adminRoleID, orgID); err != nil {
		return err
	}

	return nil
}

// GetOrganization retrieves an organization by ID
func (uc *organizationUseCase) GetOrganization(ctx context.Context, id string) (*model.OrganizationResponse, error) {
	org, err := uc.OrgRepo.FindByID(ctx, id)
	if err != nil {
		uc.Log.WithContext(ctx).Errorf("Failed to find organization: %v", err)
		return nil, exception.ErrInternalServer
	}
	if org == nil {
		return nil, exception.ErrNotFound
	}
	return converter.OrganizationToResponse(org), nil
}

// GetOrganizationBySlug retrieves an organization by slug
func (uc *organizationUseCase) GetOrganizationBySlug(ctx context.Context, slug string) (*model.OrganizationResponse, error) {
	org, err := uc.OrgRepo.FindBySlug(ctx, slug)
	if err != nil {
		uc.Log.WithContext(ctx).Errorf("Failed to find organization by slug: %v", err)
		return nil, exception.ErrInternalServer
	}
	if org == nil {
		return nil, exception.ErrNotFound
	}
	return converter.OrganizationToResponse(org), nil
}

// UpdateOrganization updates an organization's details
func (uc *organizationUseCase) UpdateOrganization(ctx context.Context, id string, request *model.UpdateOrganizationRequest) (*model.OrganizationResponse, error) {
	var response *model.OrganizationResponse

	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		org, err := uc.authorizeOrganizationManagement(txCtx, id)
		if err != nil {
			return err
		}

		// Update fields
		if request.Name != "" {
			org.Name = request.Name
		}
		if request.Settings != nil {
			org.Settings = request.Settings
		}
		if request.Status != "" {
			org.Status = request.Status
		}

		if err := uc.OrgRepo.Update(txCtx, org); err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to update organization: %v", err)
			return exception.ErrInternalServer
		}

		response = converter.OrganizationToResponse(org)
		return nil
	})

	return response, err
}

func (uc *organizationUseCase) authorizeOrganizationManagement(ctx context.Context, orgID string) (*entity.Organization, error) {
	org, err := uc.OrgRepo.FindByID(ctx, orgID)
	if err != nil {
		uc.Log.WithContext(ctx).Errorf("Failed to find organization: %v", err)
		return nil, exception.ErrInternalServer
	}
	if org == nil {
		return nil, exception.ErrNotFound
	}

	actorUserID, ok := actorUserIDFromContext(ctx)
	if !ok {
		return nil, exception.ErrForbidden
	}

	if org.OwnerID == actorUserID {
		return org, nil
	}

	isMember, err := uc.MemberRepo.CheckMembership(ctx, orgID, actorUserID)
	if err != nil {
		uc.Log.WithContext(ctx).Errorf("Failed to validate actor membership: %v", err)
		return nil, exception.ErrInternalServer
	}
	if !isMember {
		return nil, exception.ErrForbidden
	}

	roleID, err := uc.MemberRepo.GetMemberRole(ctx, orgID, actorUserID)
	if err != nil {
		uc.Log.WithContext(ctx).Errorf("Failed to get actor organization role: %v", err)
		return nil, exception.ErrInternalServer
	}

	if roleID != adminRoleID && roleID != DefaultOwnerRoleID {
		return nil, exception.ErrForbidden
	}

	return org, nil
}

// GetUserOrganizations retrieves all organizations a user is a member of
func (uc *organizationUseCase) GetUserOrganizations(ctx context.Context, userID string) (*model.UserOrganizationsResponse, error) {
	orgs, err := uc.OrgRepo.FindUserOrganizations(ctx, userID)
	if err != nil {
		uc.Log.WithContext(ctx).Errorf("Failed to find user organizations: %v", err)
		return nil, exception.ErrInternalServer
	}

	return &model.UserOrganizationsResponse{
		Organizations: converter.OrganizationsToResponse(orgs),
		Total:         len(orgs),
	}, nil
}

// DeleteOrganization deletes an organization (owner only)
func (uc *organizationUseCase) DeleteOrganization(ctx context.Context, id string, userID string) error {
	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		org, err := uc.OrgRepo.FindByID(txCtx, id)
		if err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to find organization: %v", err)
			return exception.ErrInternalServer
		}
		if org == nil {
			return exception.ErrNotFound
		}

		// Only owner can delete
		if org.OwnerID != userID {
			uc.Log.WithContext(txCtx).Warnf("User %s is not the owner of org %s", userID, id)
			return exception.ErrForbidden
		}

		if err := uc.OrgRepo.Delete(txCtx, id); err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to delete organization: %v", err)
			return exception.ErrInternalServer
		}

		return nil
	})
	if err != nil {
		return err
	}

	if uc.OrgReader != nil {
		if invalidateErr := uc.OrgReader.InvalidateOrganizationCache(ctx, id); invalidateErr != nil {
			uc.Log.WithContext(ctx).WithError(invalidateErr).Warn("Failed to invalidate organization membership cache after soft delete")
		}
	}

	return nil
}

// RestoreOrganization restores a soft-deleted organization.
func (uc *organizationUseCase) RestoreOrganization(ctx context.Context, id string) (*model.OrganizationResponse, error) {
	role, ok := actorRoleFromContext(ctx)
	if !ok || role != superAdminRoleID {
		return nil, exception.ErrForbidden
	}

	restoreCtx := database.SetAllowDeletedOrganizations(ctx, true)

	var response *model.OrganizationResponse
	err := uc.TM.WithinTransaction(restoreCtx, func(txCtx context.Context) error {
		org, err := uc.OrgRepo.FindByID(txCtx, id)
		if err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to find organization for restore: %v", err)
			return exception.ErrInternalServer
		}
		if org == nil {
			return exception.ErrNotFound
		}

		if org.DeletedAt == 0 {
			response = converter.OrganizationToResponse(org)
			return nil
		}

		if err := uc.OrgRepo.Restore(txCtx, id); err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to restore organization: %v", err)
			return exception.ErrInternalServer
		}

		org.DeletedAt = 0
		response = converter.OrganizationToResponse(org)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if uc.OrgReader != nil {
		if invalidateErr := uc.OrgReader.InvalidateOrganizationCache(ctx, id); invalidateErr != nil {
			uc.Log.WithContext(ctx).WithError(invalidateErr).Warn("Failed to invalidate organization membership cache after restore")
		}
	}

	return response, nil
}

// HardDeleteOrganization permanently deletes an already soft-deleted organization.
func (uc *organizationUseCase) HardDeleteOrganization(ctx context.Context, id string) error {
	role, ok := actorRoleFromContext(ctx)
	if !ok || role != superAdminRoleID {
		return exception.ErrForbidden
	}

	deleteCtx := database.SetAllowDeletedOrganizations(ctx, true)
	err := uc.TM.WithinTransaction(deleteCtx, func(txCtx context.Context) error {
		org, err := uc.OrgRepo.FindByID(txCtx, id)
		if err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to find organization for hard delete: %v", err)
			return exception.ErrInternalServer
		}
		if org == nil {
			return exception.ErrNotFound
		}
		if org.DeletedAt == 0 {
			return exception.ErrBadRequest
		}

		if err := uc.OrgRepo.HardDelete(txCtx, id); err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to hard delete organization: %v", err)
			return exception.ErrInternalServer
		}
		return nil
	})
	if err != nil {
		return err
	}

	if uc.OrgReader != nil {
		if invalidateErr := uc.OrgReader.InvalidateOrganizationCache(ctx, id); invalidateErr != nil {
			uc.Log.WithContext(ctx).WithError(invalidateErr).Warn("Failed to invalidate organization membership cache after hard delete")
		}
	}

	return nil
}
