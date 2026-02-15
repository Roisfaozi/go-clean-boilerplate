package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model/converter"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	permissionUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	pkgUtil "github.com/Roisfaozi/go-clean-boilerplate/pkg"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	wsPkg "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type organizationMemberUseCase struct {
	log             *logrus.Logger
	tm              tx.WithTransactionManager
	memberRepo      repository.OrganizationMemberRepository
	orgRepo         repository.OrganizationRepository
	invitationRepo  repository.InvitationRepository
	userRepo        userRepo.UserRepository
	taskDistributor worker.TaskDistributor
	enforcer        permissionUseCase.IEnforcer
	presence        PresenceReader
}

// PresenceReader defines the read-only presence operations to avoid circular dependency
type PresenceReader interface {
	GetOnlineUsers(ctx context.Context, orgID string) ([]wsPkg.PresenceUser, error)
}

// NewOrganizationMemberUseCase creates a new OrganizationMemberUseCase
func NewOrganizationMemberUseCase(
	log *logrus.Logger,
	tm tx.WithTransactionManager,
	memberRepo repository.OrganizationMemberRepository,
	orgRepo repository.OrganizationRepository,
	invitationRepo repository.InvitationRepository,
	userRepo userRepo.UserRepository,
	taskDistributor worker.TaskDistributor,
	enforcer permissionUseCase.IEnforcer,
	presence PresenceReader,
) OrganizationMemberUseCase {
	return &organizationMemberUseCase{
		log:             log,
		tm:              tm,
		memberRepo:      memberRepo,
		orgRepo:         orgRepo,
		invitationRepo:  invitationRepo,
		userRepo:        userRepo,
		taskDistributor: taskDistributor,
		enforcer:        enforcer,
		presence:        presence,
	}
}

// InviteMember invites a user to an organization
func (uc *organizationMemberUseCase) InviteMember(ctx context.Context, orgID string, request *model.InviteMemberRequest) (*model.MemberResponse, error) {
	var result *model.MemberResponse

	err := uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// 1. Check if organization exists
		org, err := uc.orgRepo.FindByID(txCtx, orgID)
		if err != nil {
			return err
		}
		if org == nil {
			return exception.ErrNotFound
		}

		// 2. Find user or create shadow user
		var targetUser *userEntity.User
		targetUser, err = uc.userRepo.FindByEmail(txCtx, request.Email)
		if err != nil && err.Error() != "user not found" {
			return err
		}

		if targetUser == nil {
			// Create Shadow User
			shadowUser := &userEntity.User{
				ID:        uuid.New().String(),
				Email:     request.Email,
				Username:  request.Email, // Use email as username for shadow users
				Status:    "invited",     // Special status for shadow users
				CreatedAt: time.Now().UnixMilli(),
				UpdatedAt: time.Now().UnixMilli(),
			}
			if err := uc.userRepo.Create(txCtx, shadowUser); err != nil {
				return err
			}
			targetUser = shadowUser
		}

		// 3. Check if user is already a member
		isMember, err := uc.memberRepo.CheckMembership(txCtx, orgID, targetUser.ID)
		if err != nil {
			return err
		}
		if isMember {
			return exception.ErrConflict
		}

		// 4. Create new member logic (if not already existing pending invite?)
		// For now simple implementation: Add member with status 'invited'
		member := &entity.OrganizationMember{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			UserID:         targetUser.ID,
			RoleID:         request.RoleID,
			Status:         entity.MemberStatusInvited,
		}

		// Check if member record already exists (e.g. previous invite)
		existingStatus, err := uc.memberRepo.GetMemberStatus(txCtx, orgID, targetUser.ID)
		if err != nil {
			return err
		}

		if existingStatus == "" {
			if err := uc.memberRepo.AddMember(txCtx, member); err != nil {
				return err
			}
		}

		// 5. Generate Invitation Token
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			return err
		}
		tokenString := hex.EncodeToString(tokenBytes)

		invitation := &entity.InvitationToken{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			Email:          targetUser.Email,
			Token:          tokenString,
			Role:           request.RoleID,
			ExpiresAt:      time.Now().Add(48 * time.Hour).UnixMilli(), // 48 hours expiry
			CreatedAt:      time.Now().UnixMilli(),
		}

		// Clean up old invitations for this email/org
		if err := uc.invitationRepo.DeleteByEmailAndOrg(txCtx, targetUser.Email, orgID); err != nil {
			return err
		}

		if err := uc.invitationRepo.Create(txCtx, invitation); err != nil {
			return err
		}

		// 6. Send Invitation Email (Async)
		// Construct email payload
		payload := &tasks.SendEmailPayload{
			To:      targetUser.Email,
			Subject: "Invitation to join organization",
			Body:    "You have been invited to join " + org.Name + ". Click here to accept: " + "http://localhost:3000/accept-invite?token=" + tokenString, // TODO: Use config for URL
		}

		if err := uc.taskDistributor.DistributeTaskSendEmail(txCtx, payload); err != nil {
			uc.log.WithError(err).Warn("Failed to queue invitation email")
			// Don't fail the transaction if email fails, user can resend
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

		// Update Casbin grouping policy if role changed
		if request.RoleID != "" && uc.enforcer != nil {
			// Remove all existing roles for this user in this org
			if _, err := uc.enforcer.RemoveFilteredGroupingPolicy(0, userID, "", orgID); err != nil {
				uc.log.WithError(err).Error("Failed to remove old Casbin grouping policy")
			}
			// Add new role
			if _, err := uc.enforcer.AddGroupingPolicy(userID, request.RoleID, orgID); err != nil {
				uc.log.WithError(err).Error("Failed to add new Casbin grouping policy")
				return exception.ErrInternalServer
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

// AcceptInvitation accepts an invitation and activates the user if needed
func (uc *organizationMemberUseCase) AcceptInvitation(ctx context.Context, request *model.AcceptInvitationRequest) error {
	return uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// 1. Find Invitation
		invitation, err := uc.invitationRepo.FindByToken(txCtx, request.Token)
		if err != nil {
			return exception.ErrUnauthorized // Token invalid or not found
		}
		if invitation == nil {
			return exception.ErrUnauthorized
		}

		// 2. Check Expiry
		if invitation.ExpiresAt < time.Now().UnixMilli() {
			return exception.ErrUnauthorized // Token expired
		}

		// 3. Find User
		user, err := uc.userRepo.FindByEmail(txCtx, invitation.Email)
		if err != nil {
			return err
		}
		if user == nil {
			// Should not happen if invitation exists, as we create shadow user.
			return exception.ErrNotFound
		}

		// 4. Update User if it's a shadow user (status "invited")
		// NOTE: Assuming "invited" status. If user is already active, we just link to org.
		if user.Status == "invited" {
			if request.Password == "" {
				return exception.ErrBadRequest // Password required for new user
			}

			hash, err := pkgUtil.HashPassword(request.Password)
			if err != nil {
				return err
			}

			user.Password = hash
			user.Status = "active"
			if request.Name != "" {
				user.Name = pkgUtil.SanitizeString(request.Name)
			} else {
				user.Name = user.Email // Default name
			}
			user.EmailVerifiedAt = new(int64)
			*user.EmailVerifiedAt = time.Now().UnixMilli()

			if err := uc.userRepo.Update(txCtx, user); err != nil {
				return err
			}
		}

		// 5. Update Organization Member Status
		// We need to find the member record first.
		memberStatus, err := uc.memberRepo.GetMemberStatus(txCtx, invitation.OrganizationID, user.ID)
		if err != nil {
			return err
		}

		switch memberStatus {
		case entity.MemberStatusInvited:
			if err := uc.memberRepo.UpdateMemberStatus(txCtx, invitation.OrganizationID, user.ID, entity.MemberStatusActive); err != nil {
				return err
			}
		case "":
			member := &entity.OrganizationMember{
				ID:             uuid.New().String(),
				OrganizationID: invitation.OrganizationID,
				UserID:         user.ID,
				RoleID:         invitation.Role,
				Status:         entity.MemberStatusActive,
			}
			if err := uc.memberRepo.AddMember(txCtx, member); err != nil {
				return err
			}
		}
		// If already active, do nothing (idempotent).

		// Add Casbin Grouping Policy for new active member
		if uc.enforcer != nil {
			// Find member to get role if it was case entity.MemberStatusInvited
			roleID := invitation.Role
			if _, err := uc.enforcer.AddGroupingPolicy(user.ID, roleID, invitation.OrganizationID); err != nil {
				uc.log.WithError(err).Error("Failed to add Casbin grouping policy on accept")
				return exception.ErrInternalServer
			}
		}

		// 6. Delete Invitation
		if err := uc.invitationRepo.Delete(txCtx, invitation.ID); err != nil {
			return err
		}

		return nil
	})
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

		// Remove Casbin grouping policy
		if uc.enforcer != nil {
			if _, err := uc.enforcer.RemoveFilteredGroupingPolicy(0, userID, "", orgID); err != nil {
				uc.log.WithError(err).Error("Failed to remove Casbin grouping policy on member removal")
			}
		}

		return uc.memberRepo.RemoveMember(txCtx, orgID, userID)
	})
}

// GetPresence retrieves online members of an organization
func (uc *organizationMemberUseCase) GetPresence(ctx context.Context, orgID string) ([]interface{}, error) {
	users, err := uc.presence.GetOnlineUsers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(users))
	for i, u := range users {
		result[i] = u
	}
	return result, nil
}
