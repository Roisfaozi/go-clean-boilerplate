package test

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestRoleContract_IDVsName_AcceptInvitation verifies that when accepting invitation:
// - DB stores role.ID
// - Casbin grouping policy receives role.Name (not role.ID)
func TestRoleContract_IDVsName_AcceptInvitation(t *testing.T) {
	deps, uc := setupMemberTest()
	ctx := context.Background()

	roleUUID := "018f3a5b-7c9d-7000-8000-111111111111"
	roleName := "custom_manager"
	orgID := "org-test-123"

	req := &model.AcceptInvitationRequest{
		Token:    "valid-token",
		Password: "password123",
		Name:     "Accepted User",
	}

	inv := &entity.InvitationToken{
		ID:             "inv-123",
		OrganizationID: orgID,
		Email:          "invitee@example.com",
		RoleID:         roleUUID, // Stored as role ID
		ExpiresAt:      time.Now().Add(1 * time.Hour).UnixMilli(),
	}

	user := &userEntity.User{
		ID:     "user-invitee-1",
		Email:  "invitee@example.com",
		Status: "invited",
	}

	mockRole := &roleEntity.Role{
		ID:             roleUUID,
		Name:           roleName,
		OrganizationID: &orgID,
	}

	deps.RoleRepo.On("FindOrganizationRoleByID", mock.Anything, orgID, roleUUID).Return(mockRole, nil)

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
	deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
	deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
	deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return(entity.MemberStatusInvited, nil)
	deps.MemberRepo.On("UpdateMemberStatus", ctx, orgID, user.ID, entity.MemberStatusActive).Return(nil)

	deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
	// Casbin MUST be called with roleName ("custom_manager"), NOT roleUUID ("018f3a5b-...")
	deps.Enforcer.On("AddGroupingPolicy", mock.MatchedBy(func(params []interface{}) bool {
		return len(params) == 3 && params[0] == user.ID && params[1] == roleName && params[2] == orgID
	})).Return(true, nil)
	deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(nil)

	err := uc.AcceptInvitation(ctx, req)
	require.NoError(t, err)
	deps.Enforcer.AssertCalled(t, "AddGroupingPolicy", mock.MatchedBy(func(params []interface{}) bool {
		return len(params) == 3 && params[0] == user.ID && params[1] == roleName && params[2] == orgID
	}))
}

func TestRoleContract_InviteOwnerRole_Forbidden(t *testing.T) {
	deps, uc := setupMemberTest()
	actorID := "owner-1"
	ctx := usecase.WithActorUserID(context.Background(), actorID)
	orgID := "org-test-123"

	req := &model.InviteMemberRequest{
		Email:  "someone@example.com",
		RoleID: usecase.DefaultOwnerRoleName,
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	deps.OrgRepo.On("FindByID", ctx, orgID).Return(&entity.Organization{ID: orgID, OwnerID: actorID}, nil)

	_, err := uc.InviteMember(ctx, orgID, req)
	require.ErrorIs(t, err, exception.ErrForbidden)
}

func TestRoleContract_CrossTenantRole_BadRequest(t *testing.T) {
	deps, uc := setupMemberTest()
	actorID := "owner-1"
	ctx := usecase.WithActorUserID(context.Background(), actorID)
	orgID := "org-test-123"
	otherOrgID := "other-org"
	otherRoleID := "other-role-id"

	req := &model.InviteMemberRequest{
		Email:  "someone@example.com",
		RoleID: otherRoleID,
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	})

	deps.OrgRepo.On("FindByID", ctx, orgID).Return(&entity.Organization{ID: orgID, OwnerID: actorID}, nil)
	// Role belongs to another organization, so FindOrganizationRoleByID for orgID returns RecordNotFound
	deps.RoleRepo.On("FindOrganizationRoleByID", mock.Anything, orgID, otherRoleID).Return(nil, exception.ErrNotFound)
	_ = otherOrgID

	_, err := uc.InviteMember(ctx, orgID, req)
	require.ErrorIs(t, err, exception.ErrBadRequest)
}
