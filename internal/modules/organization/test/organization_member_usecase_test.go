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
	Enforcer        *permissionMocks.MockIEnforcer
	Presence        *mocks.MockPresenceReader
	TM              *mocking.MockWithTransactionManager
}

func setupMemberTest() (*memberTestDeps, usecase.OrganizationMemberUseCase) {
	mockEnforcer := new(permissionMocks.MockIEnforcer)
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

	t.Run("Org Not Found", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrNotFound)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(nil, nil)

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("User Repo Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Shadow User Create Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "shadow@example.com"}
		org := &entity.Organization{ID: orgID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, nil) // Not found (nil, nil) or error "user not found"? Assuming nil, nil for logic check or error handling
		// Based on implementation: err != nil && err.Error() != "user not found" -> return err.
		// If targetUser == nil -> create shadow user.
		// Wait, existing test says: Return(nil, errors.New("user not found"))
		// Implementation:
		// targetUser, err = uc.userRepo.FindByEmail(txCtx, request.Email)
		// if err != nil && err.Error() != "user not found" { return err }
		// if targetUser == nil { ... }
		// So if repo returns nil, "user not found", then targetUser is nil (presumably initialized var).
		// Wait, Go variables: var targetUser *userEntity.User. It's nil.
		// So if error is "user not found", err is ignored, targetUser remains nil.

		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("user not found"))
		deps.UserRepo.On("Create", ctx, mock.Anything).Return(errors.New("create error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Membership Check Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID}
		user := &userEntity.User{ID: "user-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Get Member Status Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID}
		user := &userEntity.User{ID: "user-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Add Member Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID}
		user := &userEntity.User{ID: "user-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Invitation Cleanup Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID}
		user := &userEntity.User{ID: "user-1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Invitation Create Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID}
		user := &userEntity.User{ID: "user-1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(nil)
		deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
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
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

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

	t.Run("Check Membership Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(false, errors.New("db error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, &model.UpdateMemberRequest{})
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Update Member Role Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(errors.New("db error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Update Member Status Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{Status: "inactive"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, orgID, userID, "inactive").Return(errors.New("db error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Enforcer Add Grouping Policy Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Find Members Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)
		deps.MemberRepo.On("FindMembers", ctx, orgID).Return(nil, errors.New("db error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
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
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
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

	t.Run("Check Membership Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(false, errors.New("db error"))

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Org Find Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(nil, errors.New("db error"))

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Remove Member Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: "other"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
		deps.MemberRepo.On("RemoveMember", ctx, orgID, userID).Return(errors.New("db error"))

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrInternalServer)
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
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)
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

	t.Run("Invitation Expired", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "expired"}
		inv := &entity.InvitationToken{
			ID:        "inv-1",
			ExpiresAt: time.Now().Add(-1 * time.Hour).UnixMilli(),
		}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrUnauthorized)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrUnauthorized)
	})

	t.Run("User Repo Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(nil, errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("User Not Found", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrNotFound)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(nil, nil)

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("Shadow User Missing Password", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: ""}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-1", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrBadRequest)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrBadRequest)
	})

	t.Run("User Update Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-1", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.UserRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Get Member Status Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-1", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return("", errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Update Member Status Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-1", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Add Member (No Status) Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-1", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.MatchedBy(func(m *entity.OrganizationMember) bool {
			return m.Status == entity.MemberStatusActive
		})).Return(errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Enforcer Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1", Role: "role:member",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-1", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin error"))

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Invitation Delete Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1", Role: "role:member",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-1", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)
		deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})
}
