package usecase

import (
	"context"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model/converter"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type organizationMemberUseCase struct {
	log        *logrus.Logger
	tm         tx.WithTransactionManager
	memberRepo repository.OrganizationMemberRepository
	orgRepo    repository.OrganizationRepository
}

// NewOrganizationMemberUseCase creates a new OrganizationMemberUseCase
func NewOrganizationMemberUseCase(
	log *logrus.Logger,
	tm tx.WithTransactionManager,
	memberRepo repository.OrganizationMemberRepository,
	orgRepo repository.OrganizationRepository,
) OrganizationMemberUseCase {
	return &organizationMemberUseCase{
		log:        log,
		tm:         tm,
		memberRepo: memberRepo,
		orgRepo:    orgRepo,
	}
}

// InviteMember invites a user to an organization
func (uc *organizationMemberUseCase) InviteMember(ctx context.Context, orgID string, request *model.InviteMemberRequest) (*model.MemberResponse, error) {
	var result *model.MemberResponse

	err := uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Check if organization exists
		org, err := uc.orgRepo.FindByID(txCtx, orgID)
		if err != nil {
			return err
		}
		if org == nil {
			return exception.ErrNotFound
		}

		// Check if user is already a member
		isMember, err := uc.memberRepo.CheckMembership(txCtx, orgID, request.UserID)
		if err != nil {
			return err
		}
		if isMember {
			return exception.ErrConflict
		}

		// Create new member
		member := &entity.OrganizationMember{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			UserID:         request.UserID,
			RoleID:         request.RoleID,
			Status:         entity.MemberStatusInvited,
		}

		if err := uc.memberRepo.AddMember(txCtx, member); err != nil {
			return err
		}

		result = converter.MemberToResponse(member)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetMembers retrieves all members of an organization
func (uc *organizationMemberUseCase) GetMembers(ctx context.Context, orgID string) ([]model.MemberResponse, error) {
	members, err := uc.memberRepo.FindMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	result := make([]model.MemberResponse, 0, len(members))
	for _, m := range members {
		result = append(result, *converter.MemberToResponse(m))
	}

	return result, nil
}

// UpdateMember updates a member's role or status
func (uc *organizationMemberUseCase) UpdateMember(ctx context.Context, orgID, userID string, request *model.UpdateMemberRequest) (*model.MemberResponse, error) {
	var result *model.MemberResponse

	err := uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Check if member exists
		isMember, err := uc.memberRepo.CheckMembership(txCtx, orgID, userID)
		if err != nil {
			return err
		}
		if !isMember {
			return exception.ErrNotFound
		}

		// Update role if provided
		if request.RoleID != "" {
			if err := uc.memberRepo.UpdateMemberRole(txCtx, orgID, userID, request.RoleID); err != nil {
				return err
			}
		}

		// Update status if provided
		if request.Status != "" {
			if err := uc.memberRepo.UpdateMemberStatus(txCtx, orgID, userID, request.Status); err != nil {
				return err
			}
		}

		// Fetch updated member data
		members, err := uc.memberRepo.FindMembers(txCtx, orgID)
		if err != nil {
			return err
		}
		for _, m := range members {
			if m.UserID == userID {
				result = converter.MemberToResponse(m)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// RemoveMember removes a member from an organization
func (uc *organizationMemberUseCase) RemoveMember(ctx context.Context, orgID, userID string) error {
	return uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Check if member exists
		isMember, err := uc.memberRepo.CheckMembership(txCtx, orgID, userID)
		if err != nil {
			return err
		}
		if !isMember {
			return exception.ErrNotFound
		}

		// Prevent removing owner
		org, err := uc.orgRepo.FindByID(txCtx, orgID)
		if err != nil {
			return err
		}
		if org != nil && org.OwnerID == userID {
			return exception.ErrForbidden
		}

		return uc.memberRepo.RemoveMember(txCtx, orgID, userID)
	})
}
