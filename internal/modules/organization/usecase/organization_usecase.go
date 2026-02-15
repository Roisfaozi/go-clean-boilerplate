package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model/converter"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultOwnerRoleID is the default role assigned to organization owners
	DefaultOwnerRoleID = "role:org-owner"
)

type organizationUseCase struct {
	Log        *logrus.Logger
	TM         tx.WithTransactionManager
	OrgRepo    repository.OrganizationRepository
	MemberRepo repository.OrganizationMemberRepository
	Enforcer   permissionUseCase.IEnforcer
}

// NewOrganizationUseCase creates a new OrganizationUseCase instance
func NewOrganizationUseCase(
	log *logrus.Logger,
	tm tx.WithTransactionManager,
	orgRepo repository.OrganizationRepository,
	memberRepo repository.OrganizationMemberRepository,
	enforcer permissionUseCase.IEnforcer,
) OrganizationUseCase {
	return &organizationUseCase{
		Log:        log,
		TM:         tm,
		OrgRepo:    orgRepo,
		MemberRepo: memberRepo,
		Enforcer:   enforcer,
	}
}

// CreateOrganization creates a new organization with the current user as owner
func (uc *organizationUseCase) CreateOrganization(ctx context.Context, userID string, request *model.CreateOrganizationRequest) (*model.OrganizationResponse, error) {
	request.Name = pkg.SanitizeString(request.Name)

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
			if _, err := uc.Enforcer.AddGroupingPolicy(userID, DefaultOwnerRoleID, org.ID); err != nil {
				uc.Log.WithContext(txCtx).Errorf("Failed to add Casbin grouping policy: %v", err)
				return exception.ErrInternalServer
			}
		}

		response = converter.OrganizationToResponse(org)
		return nil
	})

	return response, err
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
	if request.Name != "" {
		request.Name = pkg.SanitizeString(request.Name)
	}

	var response *model.OrganizationResponse

	err := uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
		org, err := uc.OrgRepo.FindByID(txCtx, id)
		if err != nil {
			uc.Log.WithContext(txCtx).Errorf("Failed to find organization: %v", err)
			return exception.ErrInternalServer
		}
		if org == nil {
			return exception.ErrNotFound
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
	return uc.TM.WithinTransaction(ctx, func(txCtx context.Context) error {
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
}
