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
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
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
	orgReader       IOrganizationReader
	frontendBaseURL string
	roleRepo        roleRepo.RoleRepository
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
	orgReader IOrganizationReader,
	frontendBaseURL string,
	roleRepo roleRepo.RoleRepository,
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
		orgReader:       orgReader,
		frontendBaseURL: frontendBaseURL,
		roleRepo:        roleRepo,
	}
}

func (uc *organizationMemberUseCase) resolveOrganizationRole(ctx context.Context, orgID, roleID string) (*roleEntity.Role, error) {
	if uc.roleRepo == nil {
		return nil, exception.ErrInternalServer
	}

	role, err := uc.roleRepo.FindOrganizationRoleByID(ctx, orgID, roleID)
	if err == nil && role != nil {
		return role, nil
	}

	return nil, exception.ErrBadRequest
}

// InviteMember invites a user to an organization
func (uc *organizationMemberUseCase) InviteMember(ctx context.Context, orgID string, request *model.InviteMemberRequest) (*model.MemberResponse, error) {
	var result *model.MemberResponse
	var targetUserID string

	err := uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		org, _, _, err := uc.authorizeMemberManagement(txCtx, orgID)
		if err != nil {
			return err
		}
		if request.RoleID == DefaultOwnerRoleName {
			return exception.ErrForbidden
		}

		targetRole, err := uc.resolveOrganizationRole(txCtx, orgID, request.RoleID)
		if err != nil {
			return err
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
		targetUserID = targetUser.ID

		// 3. Check if user is already an active member
		existingStatus, err := uc.memberRepo.GetMemberStatus(txCtx, orgID, targetUser.ID)
		if err != nil {
			return err
		}
		if existingStatus == entity.MemberStatusActive {
			return exception.ErrConflict
		}

		// 4. Create or update member record for invited status
		member := &entity.OrganizationMember{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			UserID:         targetUser.ID,
			RoleID:         targetRole.ID,
			Status:         entity.MemberStatusInvited,
		}

		if existingStatus == "" {
			if err := uc.memberRepo.AddMember(txCtx, member); err != nil {
				return err
			}
		} else if existingStatus == entity.MemberStatusInvited {
			if err := uc.memberRepo.UpdateMemberRole(txCtx, orgID, targetUser.ID, targetRole.ID); err != nil {
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
			RoleID:         targetRole.ID,
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
			Body:    "You have been invited to join " + org.Name + ". Click here to accept: " + uc.frontendBaseURL + "/accept-invite?token=" + tokenString,
		}

		if uc.taskDistributor != nil {
			if err := uc.taskDistributor.DistributeTaskSendEmail(txCtx, payload); err != nil {
				uc.log.WithError(err).Warn("Failed to queue invitation email")
				// Don't fail the transaction if email fails, user can resend
			}
		}

		result = converter.MemberToResponse(member)
		return nil
	})

	if err != nil {
		return nil, err
	}

	uc.invalidateMembershipCache(ctx, orgID, targetUserID)

	return result, nil
}

// GetMembers retrieves all members of an organization
func (uc *organizationMemberUseCase) GetMembers(ctx context.Context, orgID string) ([]model.MemberResponse, error) {
	if _, _, _, err := uc.authorizeMemberManagement(ctx, orgID); err != nil {
		return nil, err
	}

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
		org, _, _, err := uc.authorizeMemberManagement(txCtx, orgID)
		if err != nil {
			return err
		}

		// Check if member exists and lock row
		member, err := uc.memberRepo.FindMemberForUpdate(txCtx, orgID, userID)
		if err != nil {
			return err
		}
		if member == nil {
			return exception.ErrNotFound
		}

		if org.OwnerID == userID {
			return exception.ErrForbidden
		}
		if request.RoleID == DefaultOwnerRoleName {
			return exception.ErrForbidden
		}

		var newRole *roleEntity.Role
		if request.RoleID != "" {
			newRole, err = uc.resolveOrganizationRole(txCtx, orgID, request.RoleID)
			if err != nil {
				return err
			}
		}

		oldRoleName, _ := uc.memberRepo.GetMemberRoleName(txCtx, orgID, userID)
		if oldRoleName == "" {
			oldRoleName = member.RoleID
		}

		// Update role if provided
		if newRole != nil {
			if err := uc.memberRepo.UpdateMemberRole(txCtx, orgID, userID, newRole.ID); err != nil {
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
		if newRole != nil && uc.enforcer != nil {
			enf := uc.enforcer.WithContext(txCtx)
			if _, err := enf.RemoveFilteredGroupingPolicy(0, userID, oldRoleName, orgID); err != nil {
				uc.log.WithError(err).Error("Failed to remove old Casbin grouping policy")
				return exception.ErrInternalServer
			}
			if _, err := enf.AddGroupingPolicy(userID, newRole.Name, orgID); err != nil {
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

	uc.reloadPolicyAfterCommit(ctx)
	uc.invalidateMembershipCache(ctx, orgID, userID)

	return result, nil
}

// AcceptInvitation accepts an invitation and activates the user if needed
func (uc *organizationMemberUseCase) AcceptInvitation(ctx context.Context, request *model.AcceptInvitationRequest) error {
	var orgID string
	var userID string

	err := uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
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

		// 5. Resolve role to get role.Name for Casbin and role.ID for DB
		role, err := uc.resolveOrganizationRole(txCtx, invitation.OrganizationID, invitation.RoleID)
		if err != nil {
			uc.log.WithContext(txCtx).Warnf("Failed to resolve organization role %s for invitation: %v", invitation.RoleID, err)
			return exception.ErrUnauthorized
		}

		memberStatus, err := uc.memberRepo.GetMemberStatus(txCtx, invitation.OrganizationID, user.ID)
		if err != nil {
			return err
		}

		switch memberStatus {
		case entity.MemberStatusInvited:
			if err := uc.memberRepo.UpdateMemberStatus(txCtx, invitation.OrganizationID, user.ID, entity.MemberStatusActive); err != nil {
				return err
			}
			if uc.enforcer != nil {
				if _, err := uc.enforcer.WithContext(txCtx).AddGroupingPolicy(user.ID, role.Name, invitation.OrganizationID); err != nil {
					uc.log.WithError(err).Error("Failed to add Casbin grouping policy on accept")
					return exception.ErrInternalServer
				}
			}
		case "":
			member := &entity.OrganizationMember{
				ID:             uuid.New().String(),
				OrganizationID: invitation.OrganizationID,
				UserID:         user.ID,
				RoleID:         role.ID,
				Status:         entity.MemberStatusActive,
			}
			if err := uc.memberRepo.AddMember(txCtx, member); err != nil {
				return err
			}
			if uc.enforcer != nil {
				if _, err := uc.enforcer.WithContext(txCtx).AddGroupingPolicy(user.ID, role.Name, invitation.OrganizationID); err != nil {
					uc.log.WithError(err).Error("Failed to add Casbin grouping policy on accept")
					return exception.ErrInternalServer
				}
			}
		}
		// If already active, do not add duplicate policy, just clean up token.
		orgID = invitation.OrganizationID
		userID = user.ID

		// 6. Delete Invitation
		if err := uc.invitationRepo.Delete(txCtx, invitation.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	uc.reloadPolicyAfterCommit(ctx)
	uc.invalidateMembershipCache(ctx, orgID, userID)
	return nil
}

// RemoveMember removes a member from an organization
func (uc *organizationMemberUseCase) RemoveMember(ctx context.Context, orgID, userID string) error {
	err := uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		org, _, _, err := uc.authorizeMemberManagement(txCtx, orgID)
		if err != nil {
			return err
		}

		// Check if member exists and lock row
		member, err := uc.memberRepo.FindMemberForUpdate(txCtx, orgID, userID)
		if err != nil {
			return err
		}
		if member == nil {
			return exception.ErrNotFound
		}

		// Prevent removing owner
		if org != nil && org.OwnerID == userID {
			return exception.ErrForbidden
		}

		oldRoleName, err := uc.memberRepo.GetMemberRoleName(txCtx, orgID, userID)
		if err != nil && oldRoleName == "" {
			oldRoleName = member.RoleID
		}

		// Remove Casbin grouping policy
		if uc.enforcer != nil {
			if _, err := uc.enforcer.WithContext(txCtx).RemoveFilteredGroupingPolicy(0, userID, oldRoleName, orgID); err != nil {
				uc.log.WithError(err).Error("Failed to remove Casbin grouping policy on member removal")
				return exception.ErrInternalServer
			}
		}

		return uc.memberRepo.RemoveMember(txCtx, orgID, userID)
	})
	if err != nil {
		return err
	}

	uc.reloadPolicyAfterCommit(ctx)
	uc.invalidateMembershipCache(ctx, orgID, userID)
	return nil
}

func (uc *organizationMemberUseCase) authorizeMemberManagement(ctx context.Context, orgID string) (*entity.Organization, string, bool, error) {
	org, err := uc.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return nil, "", false, exception.ErrInternalServer
	}
	if org == nil {
		return nil, "", false, exception.ErrNotFound
	}

	actorUserID, ok := actorUserIDFromContext(ctx)
	if !ok {
		// Fail closed: if we don't know who the actor is, we cannot authorize the action.
		return nil, "", false, exception.ErrForbidden
	}

	if org.OwnerID == actorUserID {
		return org, actorUserID, true, nil
	}

	isMember, err := uc.memberRepo.CheckMembership(ctx, orgID, actorUserID)
	if err != nil {
		return nil, "", false, exception.ErrInternalServer
	}
	if !isMember {
		return nil, "", false, exception.ErrForbidden
	}

	roleName, err := uc.memberRepo.GetMemberRoleName(ctx, orgID, actorUserID)
	if err != nil {
		return nil, "", false, exception.ErrInternalServer
	}
	if roleName != adminRoleID && roleName != DefaultOwnerRoleName {
		return nil, "", false, exception.ErrForbidden
	}

	return org, actorUserID, false, nil
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

func (uc *organizationMemberUseCase) reloadPolicyAfterCommit(ctx context.Context) {
	if uc.enforcer == nil {
		return
	}
	if err := uc.enforcer.LoadPolicy(); err != nil {
		uc.log.WithContext(ctx).WithError(err).Warn("Failed to reload Casbin policy after member operation")
	}
}

func (uc *organizationMemberUseCase) invalidateMembershipCache(ctx context.Context, orgID, userID string) {
	if uc.orgReader == nil || orgID == "" || userID == "" {
		return
	}
	if err := uc.orgReader.InvalidateMembershipCache(ctx, orgID, userID); err != nil {
		uc.log.WithContext(ctx).WithError(err).Warn("Failed to invalidate membership cache")
	}
}
