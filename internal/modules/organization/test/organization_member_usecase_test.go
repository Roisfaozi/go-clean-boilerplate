package test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	permissionMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/worker/tasks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	wsPkg "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type memberTestDeps struct {
	MemberRepo      *mocks.MockOrganizationMemberRepository
	OrgRepo         *mocks.MockOrganizationRepository
	InvitationRepo  *mocks.MockInvitationRepository
	UserRepo        *userMocks.MockUserRepository
	TaskDistributor *mocking.MockTaskDistributor
	Enforcer        *permissionMocks.IEnforcer
	Presence        *mocks.MockPresenceReader
	TM              *mocking.MockWithTransactionManager
}

func setupMemberTest() (*memberTestDeps, usecase.OrganizationMemberUseCase) {
	mockEnforcer := new(permissionMocks.IEnforcer)
	deps := &memberTestDeps{
		MemberRepo:      new(mocks.MockOrganizationMemberRepository),
		OrgRepo:         new(mocks.MockOrganizationRepository),
		InvitationRepo:  new(mocks.MockInvitationRepository),
		UserRepo:        new(userMocks.MockUserRepository),
		TaskDistributor: new(mocking.MockTaskDistributor),
		Enforcer:        mockEnforcer,
		Presence:        new(mocks.MockPresenceReader),
		TM:              new(mocking.MockWithTransactionManager),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)
	log.SetLevel(logrus.FatalLevel)

	uc := usecase.NewOrganizationMemberUseCase(
		log,
		deps.TM,
		deps.MemberRepo,
		deps.OrgRepo,
		deps.InvitationRepo,
		deps.UserRepo,
		deps.TaskDistributor,
		deps.Enforcer,
		deps.Presence,
		"http://localhost:3000",
	)

	return deps, uc
}

func TestOrganizationMemberUseCase_InviteMember(t *testing.T) {
	t.Run("Success - Existing User", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com", RoleID: "role:member"}
		org := &entity.Organization{ID: orgID, Name: "Org 1"}
		user := &userEntity.User{ID: "user-1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.MatchedBy(func(m *entity.OrganizationMember) bool {
			return m.UserID == user.ID && m.OrganizationID == orgID && m.Status == entity.MemberStatusInvited
		})).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(nil)
		deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(nil)
		deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.MatchedBy(func(p *tasks.SendEmailPayload) bool {
			return p.To == req.Email
		})).Return(nil)

		res, err := uc.InviteMember(ctx, orgID, req)
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, user.ID, res.UserID)
		assert.Equal(t, entity.MemberStatusInvited, res.Status)
	})

	t.Run("Success - Shadow User", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "shadow@example.com", RoleID: "role:member"}
		org := &entity.Organization{ID: orgID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("user not found"))
		deps.UserRepo.On("Create", ctx, mock.MatchedBy(func(u *userEntity.User) bool {
			return u.Email == req.Email && u.Status == "invited"
		})).Return(nil)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, mock.AnythingOfType("string")).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, mock.AnythingOfType("string")).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(nil)
		deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(nil)
		deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).Return(nil)

		res, err := uc.InviteMember(ctx, orgID, req)
		require.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("Already Member", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "member@example.com"}
		org := &entity.Organization{ID: orgID}
		user := &userEntity.User{ID: "user-1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrConflict)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(true, nil)

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrConflict)
	})
}

func TestOrganizationMemberUseCase_GetMembers(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		members := []*entity.OrganizationMember{
			{UserID: "u1", RoleID: "r1"},
			{UserID: "u2", RoleID: "r2"},
		}

		deps.MemberRepo.On("FindMembers", ctx, orgID).Return(members, nil)

		res, err := uc.GetMembers(ctx, orgID)
		require.NoError(t, err)
		assert.Len(t, res, 2)
	})
}

func TestOrganizationMemberUseCase_UpdateMember(t *testing.T) {
	t.Run("Success - Update Role", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)

		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "", orgID).Return(true, nil)
		deps.Enforcer.On("AddGroupingPolicy", userID, "new-role", orgID).Return(true, nil)

		deps.MemberRepo.On("FindMembers", ctx, orgID).Return([]*entity.OrganizationMember{{UserID: userID, RoleID: "new-role"}}, nil)

		res, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.NoError(t, err)
		assert.Equal(t, "new-role", res.RoleID)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrNotFound)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(false, nil)

		_, err := uc.UpdateMember(ctx, orgID, userID, &model.UpdateMemberRequest{})
		require.ErrorIs(t, err, exception.ErrNotFound)
	})
}

func TestOrganizationMemberUseCase_RemoveMember(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: "other-user"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "", orgID).Return(true, nil)
		deps.MemberRepo.On("RemoveMember", ctx, orgID, userID).Return(nil)

		err := uc.RemoveMember(ctx, orgID, userID)
		require.NoError(t, err)
	})

	t.Run("Forbidden - Remove Owner", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "owner"
		org := &entity.Organization{ID: orgID, OwnerID: userID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrForbidden)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrForbidden)
	})
}

func TestOrganizationMemberUseCase_GetPresence(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		presenceUsers := []wsPkg.PresenceUser{{UserID: "u1"}}

		deps.Presence.On("GetOnlineUsers", ctx, orgID).Return(presenceUsers, nil)

		res, err := uc.GetPresence(ctx, orgID)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})
}

func TestOrganizationMemberUseCase_AcceptInvitation(t *testing.T) {
	t.Run("Success - Activate New User", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass", Name: "Name"}
		inv := &entity.InvitationToken{ID: "inv-1", Email: "new@example.com", OrganizationID: "org-1", Role: "role:member", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "user-1", Email: "new@example.com", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.UserRepo.On("Update", ctx, mock.MatchedBy(func(u *userEntity.User) bool {
			return u.Status == "active" && u.Name == req.Name
		})).Return(nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("AddGroupingPolicy", user.ID, inv.Role, inv.OrganizationID).Return(true, nil)
		deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(nil)

		err := uc.AcceptInvitation(ctx, req)
		require.NoError(t, err)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "invalid"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrUnauthorized)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(nil, nil)

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrUnauthorized)
	})
}
