package test

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mocking "github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	permissionMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/test/mocks"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	roleMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
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
	OrgReader       *mocks.MockIOrganizationReader
	TM              *mocking.MockWithTransactionManager
	RoleRepo        *roleMocks.MockRoleRepository
}

func setupMemberTest(t *testing.T) (*memberTestDeps, usecase.OrganizationMemberUseCase) {
	mockEnforcer := new(permissionMocks.MockIEnforcer)
	roleRepo := roleMocks.NewMockRoleRepository(t)
	deps := &memberTestDeps{
		MemberRepo:      mocks.NewMockOrganizationMemberRepository(t),
		OrgRepo:         mocks.NewMockOrganizationRepository(t),
		InvitationRepo:  mocks.NewMockInvitationRepository(t),
		UserRepo:        userMocks.NewMockUserRepository(t),
		TaskDistributor: mocking.NewMockTaskDistributor(t),
		Enforcer:        mockEnforcer,
		Presence:        mocks.NewMockPresenceReader(t),
		OrgReader:       mocks.NewMockIOrganizationReader(t),
		TM:              mocking.NewMockWithTransactionManager(t),
		RoleRepo:        roleRepo,
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
		deps.OrgReader,
		"http://localhost:3000",
		deps.RoleRepo,
	)

	deps.OrgReader.On("InvalidateMembershipCache", mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)
	deps.Enforcer.On("LoadPolicy").Maybe().Return(nil)

	return deps, uc
}

func setupDefaultRoleMocks(deps *memberTestDeps) {
	deps.RoleRepo.On("FindOrganizationRoleByID", mock.Anything, mock.Anything, mock.Anything).Maybe().Return(func(ctx context.Context, orgID, roleID string) *roleEntity.Role {
		return &roleEntity.Role{ID: roleID, Name: roleID}
	}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, mock.Anything).Maybe().Return(func(ctx context.Context, name string) *roleEntity.Role {
		return &roleEntity.Role{ID: name, Name: name}
	}, nil)
	deps.RoleRepo.On("FindByNameInScope", mock.Anything, mock.Anything, mock.Anything).Maybe().Return(func(ctx context.Context, name string, orgID *string) *roleEntity.Role {
		return &roleEntity.Role{ID: name, Name: name}
	}, nil)
}

func TestOrganizationMemberUseCase_InviteMember(t *testing.T) {
	t.Run("Success - Existing User", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com", RoleID: "role:member"}
		org := &entity.Organization{ID: orgID, Name: "Org 1", OwnerID: actorID}
		user := &userEntity.User{ID: "user-1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil).Maybe()
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.MatchedBy(func(m *entity.OrganizationMember) bool {
			return m.UserID == user.ID && m.OrganizationID == orgID && m.Status == entity.MemberStatusInvited
		})).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(nil)
		deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(nil)
		deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.MatchedBy(func(p *tasks.SendEmailPayload) bool {
			return p.To == req.Email && strings.Contains(p.Body, "/invite/")
		})).Return(nil)

		res, err := uc.InviteMember(ctx, orgID, req)
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, user.ID, res.UserID)
		assert.Equal(t, entity.MemberStatusInvited, res.Status)
		deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, orgID, user.ID)
	})

	t.Run("Success - Shadow User", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "shadow@example.com", RoleID: "role:member"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("user not found"))
		deps.UserRepo.On("Create", ctx, mock.MatchedBy(func(u *userEntity.User) bool {
			return u.Email == req.Email && u.Status == "invited"
		})).Return(nil)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, mock.AnythingOfType("string")).Return(false, nil).Maybe()
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, mock.AnythingOfType("string")).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(nil)
		deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(nil)
		deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).Return(nil)

		res, err := uc.InviteMember(ctx, orgID, req)
		require.NoError(t, err)
		assert.NotNil(t, res)
		deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, orgID, res.UserID)
	})

	t.Run("Already Member", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "member@example.com"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		user := &userEntity.User{ID: "user-1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return(entity.MemberStatusActive, nil)

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrConflict)
	})

	t.Run("Org Not Found", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(nil, nil)

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("User Repo Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "shadow@example.com"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		user := &userEntity.User{ID: "user-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Get Member Status Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		user := &userEntity.User{ID: "user-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil).Maybe()
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Add Member Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		user := &userEntity.User{ID: "user-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil).Maybe()
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Invitation Cleanup Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		user := &userEntity.User{ID: "user-1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil).Maybe()
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(errors.New("db error"))

		_, err := uc.InviteMember(ctx, orgID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Invitation Create Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		req := &model.InviteMemberRequest{Email: "user@example.com"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		user := &userEntity.User{ID: "user-1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil).Maybe()
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		members := []*entity.OrganizationMember{
			{UserID: "u1", RoleID: "r1"},
			{UserID: "u2", RoleID: "r2"},
		}

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMembers", ctx, orgID).Return(members, nil)

		res, err := uc.GetMembers(ctx, orgID)
		require.NoError(t, err)
		assert.Len(t, res, 2)
	})
}

func TestOrganizationMemberUseCase_UpdateMember(t *testing.T) {
	t.Run("Success - Update Role", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		member := &entity.OrganizationMember{OrganizationID: orgID, UserID: userID, RoleID: "old-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(member, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, userID).Return("old-role", nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)

		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

		deps.MemberRepo.On("FindMembers", mock.Anything, orgID).Return([]*entity.OrganizationMember{{UserID: userID, RoleID: "new-role"}}, nil)

		res, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.NoError(t, err)
		assert.Equal(t, "new-role", res.RoleID)
		deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, orgID, userID)
	})

	t.Run("Not Found", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrNotFound)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(nil, nil)

		_, err := uc.UpdateMember(ctx, orgID, userID, &model.UpdateMemberRequest{})
		require.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("Check Membership Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(nil, errors.New("db error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, &model.UpdateMemberRequest{})
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Update Member Role Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		member := &entity.OrganizationMember{OrganizationID: orgID, UserID: userID, RoleID: "old-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(member, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, userID).Return("old-role", nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(errors.New("db error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Update Member Status Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{Status: "inactive"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		member := &entity.OrganizationMember{OrganizationID: orgID, UserID: userID, RoleID: "old-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(member, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, userID).Return("old-role", nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, orgID, userID, "inactive").Return(errors.New("db error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Enforcer Add Grouping Policy Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		member := &entity.OrganizationMember{OrganizationID: orgID, UserID: userID, RoleID: "old-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(member, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, userID).Return("old-role", nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(false, errors.New("casbin error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Find Members Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		member := &entity.OrganizationMember{OrganizationID: orgID, UserID: userID, RoleID: "old-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(member, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, userID).Return("old-role", nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		deps.MemberRepo.On("FindMembers", ctx, orgID).Return(nil, errors.New("db error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Forbidden - Non admin actor cannot update owner via actor context", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		ctx := usecase.WithActorUserID(context.Background(), "member-1")
		orgID := "org-1"
		userID := "owner-1"
		org := &entity.Organization{ID: orgID, OwnerID: userID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrForbidden)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, "member-1").Return(true, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, "member-1").Return("role:user", nil)

		_, err := uc.UpdateMember(ctx, orgID, userID, &model.UpdateMemberRequest{Status: entity.MemberStatusSuspended})
		require.ErrorIs(t, err, exception.ErrForbidden)
	})
}

func TestOrganizationMemberUseCase_RemoveMember(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(&entity.OrganizationMember{OrganizationID: orgID, UserID: userID, RoleID: "role:member"}, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, userID).Return("role:member", nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		deps.MemberRepo.On("RemoveMember", ctx, orgID, userID).Return(nil)

		err := uc.RemoveMember(ctx, orgID, userID)
		require.NoError(t, err)
		deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, orgID, userID)
	})

	t.Run("No-op - Member Already Gone", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(nil, nil)

		err := uc.RemoveMember(ctx, orgID, userID)
		require.NoError(t, err)
		deps.Enforcer.AssertNotCalled(t, "RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		deps.MemberRepo.AssertNotCalled(t, "RemoveMember", mock.Anything, mock.Anything, mock.Anything)
		deps.Enforcer.AssertNotCalled(t, "LoadPolicy")
		deps.OrgReader.AssertNotCalled(t, "InvalidateMembershipCache", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("No-op - FindMemberForUpdate Returns ErrNotFound", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(nil, exception.ErrNotFound)

		err := uc.RemoveMember(ctx, orgID, userID)
		require.NoError(t, err)
		deps.Enforcer.AssertNotCalled(t, "RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		deps.MemberRepo.AssertNotCalled(t, "RemoveMember", mock.Anything, mock.Anything, mock.Anything)
		deps.Enforcer.AssertNotCalled(t, "LoadPolicy")
		deps.OrgReader.AssertNotCalled(t, "InvalidateMembershipCache", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Idempotent - Repeat RemoveMember returns nil with zero extra side effects", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Twice()

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil).Twice()
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(&entity.OrganizationMember{OrganizationID: orgID, UserID: userID, RoleID: "role:member"}, nil).Once()
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, userID).Return("role:member", nil).Once()
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer).Once()
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "role:member", orgID).Return(true, nil).Once()
		deps.MemberRepo.On("RemoveMember", ctx, orgID, userID).Return(nil).Once()
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(nil, nil).Once()

		require.NoError(t, uc.RemoveMember(ctx, orgID, userID))
		require.NoError(t, uc.RemoveMember(ctx, orgID, userID))

		deps.Enforcer.AssertNumberOfCalls(t, "RemoveFilteredGroupingPolicy", 1)
		deps.MemberRepo.AssertNumberOfCalls(t, "RemoveMember", 1)
		deps.Enforcer.AssertNumberOfCalls(t, "LoadPolicy", 1)
		deps.OrgReader.AssertNumberOfCalls(t, "InvalidateMembershipCache", 1)
	})

	t.Run("Not Found - Missing Org still fails before member lookup", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "missing-org"
		userID := "user-1"

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Once()
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(nil, nil).Once()

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrNotFound)
		deps.MemberRepo.AssertNotCalled(t, "FindMemberForUpdate", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Forbidden - Remove Owner", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "owner-1"
		org := &entity.Organization{ID: orgID, OwnerID: userID}
		member := &entity.OrganizationMember{OrganizationID: orgID, UserID: userID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrForbidden)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(member, nil)

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrForbidden)
	})

	t.Run("Check Membership Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(nil, errors.New("db error"))

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Org Find Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil).Maybe()
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(nil, errors.New("db error"))

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Remove Member Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(&entity.OrganizationMember{OrganizationID: orgID, UserID: userID, RoleID: "role:member"}, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, userID).Return("role:member", nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		deps.MemberRepo.On("RemoveMember", ctx, orgID, userID).Return(errors.New("db error"))

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Forbidden - actor context role is not manager", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		ctx := usecase.WithActorUserID(context.Background(), "member-1")
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrForbidden)

		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("CheckMembership", ctx, orgID, "member-1").Return(true, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, "member-1").Return("role:user", nil)

		err := uc.RemoveMember(ctx, orgID, userID)
		require.ErrorIs(t, err, exception.ErrForbidden)
	})
}

func TestOrganizationMemberUseCase_GetPresence(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass", Name: "Name"}
		inv := &entity.InvitationToken{ID: "inv-1", Email: "new@example.com", OrganizationID: "org-1", RoleID: "role:member", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "user-1", Email: "new@example.com", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.UserRepo.On("Update", ctx, mock.MatchedBy(func(u *userEntity.User) bool {
			return u.Status == "active" && u.Name == req.Name
		})).Return(nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(nil)
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(nil)

		err := uc.AcceptInvitation(ctx, req)
		require.NoError(t, err)
		deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, inv.OrganizationID, user.ID)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: ""}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com",
			ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-1", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrBadRequest)
	})

	t.Run("User Update Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
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
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1", RoleID: "role:member",
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
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(false, errors.New("casbin error"))

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("Invitation Delete Error", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
		inv := &entity.InvitationToken{
			ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1", RoleID: "role:member",
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
		deps.Enforcer.On("AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
		deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		require.ErrorIs(t, err, exception.ErrInternalServer)
	})
}

func TestOrganizationMemberUseCase_Idempotency_InviteAndAccept(t *testing.T) {
	t.Run("re-inviting pending email replaces single active token without duplicate member", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		email := "pending@example.com"
		org := &entity.Organization{ID: orgID, Name: "Org 1", OwnerID: actorID}
		user := &userEntity.User{ID: "user-pending", Email: email, Status: "invited"}

		var activeInvites []*entity.InvitationToken
		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, email).Return(user, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil).Once()
		deps.MemberRepo.On("AddMember", ctx, mock.AnythingOfType("*entity.OrganizationMember")).Return(nil).Once()
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return(entity.MemberStatusInvited, nil).Once()
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, user.ID, "role:editor").Return(nil).Once()

		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, email, orgID).Run(func(args mock.Arguments) {
			activeInvites = nil
		}).Return(nil).Twice()
		deps.InvitationRepo.On("Create", ctx, mock.AnythingOfType("*entity.InvitationToken")).Run(func(args mock.Arguments) {
			inv := args.Get(1).(*entity.InvitationToken)
			activeInvites = append(activeInvites, inv)
		}).Return(nil).Twice()
		deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).Return(nil).Twice()

		res1, err := uc.InviteMember(ctx, orgID, &model.InviteMemberRequest{Email: email, RoleID: "role:viewer"})
		require.NoError(t, err)
		require.Equal(t, "role:viewer", res1.RoleID)
		require.Len(t, activeInvites, 1)
		firstToken := activeInvites[0].Token

		res2, err := uc.InviteMember(ctx, orgID, &model.InviteMemberRequest{Email: email, RoleID: "role:editor"})
		require.NoError(t, err)
		require.Equal(t, "role:editor", res2.RoleID)
		require.Len(t, activeInvites, 1, "re-inviting same (org, email) must keep exactly one active invite")
		assert.NotEqual(t, firstToken, activeInvites[0].Token)
		assert.Equal(t, "role:editor", activeInvites[0].RoleID)
		deps.MemberRepo.AssertNumberOfCalls(t, "AddMember", 1)
		deps.MemberRepo.AssertNumberOfCalls(t, "UpdateMemberRole", 1)
	})

	t.Run("concurrent invites for same (org, email) serialize cleanly without multiplying invites", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-race"
		email := "race@example.com"
		org := &entity.Organization{ID: orgID, Name: "Race Org", OwnerID: actorID}
		user := &userEntity.User{ID: "user-race", Email: email, Status: "invited"}

		var stateMu sync.Mutex
		memberStatus := ""
		var activeInvites []*entity.InvitationToken
		var inFlightTx atomic.Int32
		var maxConcurrentTx atomic.Int32

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			cur := inFlightTx.Add(1)
			for {
				prev := maxConcurrentTx.Load()
				if cur <= prev || maxConcurrentTx.CompareAndSwap(prev, cur) {
					break
				}
			}
			defer inFlightTx.Add(-1)
			time.Sleep(15 * time.Millisecond)
			return fn(ctx)
		})
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, email).Return(user, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return(func(context.Context, string, string) string {
			stateMu.Lock()
			defer stateMu.Unlock()
			return memberStatus
		}, nil)
		deps.MemberRepo.On("AddMember", ctx, mock.AnythingOfType("*entity.OrganizationMember")).Run(func(args mock.Arguments) {
			stateMu.Lock()
			memberStatus = entity.MemberStatusInvited
			stateMu.Unlock()
		}).Return(nil).Once()
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, email, orgID).Run(func(args mock.Arguments) {
			stateMu.Lock()
			activeInvites = nil
			stateMu.Unlock()
		}).Return(nil).Once()
		deps.InvitationRepo.On("Create", ctx, mock.AnythingOfType("*entity.InvitationToken")).Run(func(args mock.Arguments) {
			inv := args.Get(1).(*entity.InvitationToken)
			stateMu.Lock()
			activeInvites = append(activeInvites, inv)
			stateMu.Unlock()
		}).Return(nil).Once()
		deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).Return(nil).Once()

		const workers = 4
		var wg sync.WaitGroup
		wg.Add(workers)
		errs := make(chan error, workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				_, err := uc.InviteMember(ctx, orgID, &model.InviteMemberRequest{Email: email, RoleID: "role:member"})
				errs <- err
			}()
		}
		wg.Wait()
		close(errs)

		for err := range errs {
			require.NoError(t, err)
		}
		assert.Equal(t, int32(1), maxConcurrentTx.Load(), "same (org, email) invites must serialize")
		stateMu.Lock()
		assert.Len(t, activeInvites, 1, "exactly one active invite token must remain")
		stateMu.Unlock()
		deps.MemberRepo.AssertNotCalled(t, "UpdateMemberRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		deps.InvitationRepo.AssertNumberOfCalls(t, "Create", 1)
		deps.TaskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)
	})

	t.Run("repeat same-role invite within cooldown suppresses duplicate DB invite and email", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-cooldown"
		email := "cooldown@example.com"
		org := &entity.Organization{ID: orgID, Name: "Cooldown Org", OwnerID: actorID}
		user := &userEntity.User{ID: "user-cooldown", Email: email, Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Twice()
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil).Twice()
		deps.UserRepo.On("FindByEmail", ctx, email).Return(user, nil).Once()
		deps.UserRepo.On("FindByEmail", ctx, "  Cooldown@Example.com ").Return(user, nil).Once()
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil).Once()
		deps.MemberRepo.On("AddMember", ctx, mock.AnythingOfType("*entity.OrganizationMember")).Return(nil).Once()
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, email, orgID).Return(nil).Once()
		deps.InvitationRepo.On("Create", ctx, mock.AnythingOfType("*entity.InvitationToken")).Return(nil).Once()
		deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).Return(nil).Once()

		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return(entity.MemberStatusInvited, nil).Once()

		res1, err := uc.InviteMember(ctx, orgID, &model.InviteMemberRequest{Email: email, RoleID: "role:member"})
		require.NoError(t, err)
		require.NotNil(t, res1)
		assert.Equal(t, entity.MemberStatusInvited, res1.Status)
		assert.Equal(t, "role:member", res1.RoleID)

		res2, err := uc.InviteMember(ctx, orgID, &model.InviteMemberRequest{Email: "  Cooldown@Example.com ", RoleID: "role:member"})
		require.NoError(t, err)
		require.NotNil(t, res2)
		assert.Equal(t, res1.ID, res2.ID)
		assert.Equal(t, user.ID, res2.UserID)
		assert.Equal(t, "role:member", res2.RoleID)
		assert.Equal(t, entity.MemberStatusInvited, res2.Status)

		deps.MemberRepo.AssertNumberOfCalls(t, "AddMember", 1)
		deps.MemberRepo.AssertNotCalled(t, "UpdateMemberRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		deps.InvitationRepo.AssertNumberOfCalls(t, "DeleteByEmailAndOrg", 1)
		deps.InvitationRepo.AssertNumberOfCalls(t, "Create", 1)
		deps.TaskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 1)
		deps.OrgReader.AssertNumberOfCalls(t, "InvalidateMembershipCache", 1)
	})

	t.Run("repeat same-role invite after cooldown expires resends cleanly", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		configurable, ok := uc.(interface{ SetInviteResendCooldown(time.Duration) })
		require.True(t, ok, "usecase must support SetInviteResendCooldown")
		configurable.SetInviteResendCooldown(20 * time.Millisecond)

		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-cooldown-expired"
		email := "resend@example.com"
		org := &entity.Organization{ID: orgID, Name: "Resend Org", OwnerID: actorID}
		user := &userEntity.User{ID: "user-resend", Email: email, Status: "invited"}

		var activeInvites []*entity.InvitationToken
		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).Twice()
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil).Twice()
		deps.UserRepo.On("FindByEmail", ctx, email).Return(user, nil).Twice()
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil).Once()
		deps.MemberRepo.On("AddMember", ctx, mock.AnythingOfType("*entity.OrganizationMember")).Return(nil).Once()
		deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return(entity.MemberStatusInvited, nil).Once()
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, user.ID, "role:member").Return(nil).Once()
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, email, orgID).Run(func(args mock.Arguments) {
			activeInvites = nil
		}).Return(nil).Twice()
		deps.InvitationRepo.On("Create", ctx, mock.AnythingOfType("*entity.InvitationToken")).Run(func(args mock.Arguments) {
			activeInvites = append(activeInvites, args.Get(1).(*entity.InvitationToken))
		}).Return(nil).Twice()
		deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).Return(nil).Twice()

		res1, err := uc.InviteMember(ctx, orgID, &model.InviteMemberRequest{Email: email, RoleID: "role:member"})
		require.NoError(t, err)
		require.NotNil(t, res1)
		require.Len(t, activeInvites, 1)
		firstToken := activeInvites[0].Token

		time.Sleep(30 * time.Millisecond)

		res2, err := uc.InviteMember(ctx, orgID, &model.InviteMemberRequest{Email: email, RoleID: "role:member"})
		require.NoError(t, err)
		require.NotNil(t, res2)
		require.Len(t, activeInvites, 1)
		assert.NotEqual(t, firstToken, activeInvites[0].Token)

		deps.MemberRepo.AssertNumberOfCalls(t, "AddMember", 1)
		deps.MemberRepo.AssertNumberOfCalls(t, "UpdateMemberRole", 1)
		deps.InvitationRepo.AssertNumberOfCalls(t, "DeleteByEmailAndOrg", 2)
		deps.InvitationRepo.AssertNumberOfCalls(t, "Create", 2)
		deps.TaskDistributor.AssertNumberOfCalls(t, "DistributeTaskSendEmail", 2)
	})

	t.Run("accept invitation on already active member is idempotent without duplicate Casbin policy", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "stale-active-token"}
		inv := &entity.InvitationToken{
			ID:             "inv-stale",
			Email:          "active@example.com",
			OrganizationID: "org-1",
			RoleID:         "role:member",
			ExpiresAt:      time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-active", Email: "active@example.com", Status: "active"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil).Once()
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil).Once()
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusActive, nil).Once()
		deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(nil).Once()

		require.NoError(t, uc.AcceptInvitation(ctx, req))

		deps.MemberRepo.AssertNotCalled(t, "UpdateMemberStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		deps.MemberRepo.AssertNotCalled(t, "AddMember", mock.Anything, mock.Anything)
		deps.Enforcer.AssertNotCalled(t, "AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything)
		deps.Enforcer.AssertNotCalled(t, "LoadPolicy")
	})

	t.Run("concurrent accept invitation with same token executes once", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "concurrent-accept-token", Password: "Password123!", Name: "New User"}
		inv := &entity.InvitationToken{
			ID:             "inv-concurrent",
			Email:          "concurrent@example.com",
			OrganizationID: "org-1",
			RoleID:         "role:member",
			ExpiresAt:      time.Now().Add(1 * time.Hour).UnixMilli(),
		}
		user := &userEntity.User{ID: "user-concurrent", Email: "concurrent@example.com", Status: "invited"}

		started := make(chan struct{})
		release := make(chan struct{})

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			close(started)
			<-release
			return fn(ctx)
		}).Once()
		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil).Once()
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil).Once()
		deps.UserRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(nil).Once()
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil).Once()
		deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(nil).Once()
		deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer).Once()
		deps.Enforcer.On("AddGroupingPolicy", user.ID, "role:member", inv.OrganizationID).Return(true, nil).Once()
		deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(nil).Once()

		var wg sync.WaitGroup
		errs := make(chan error, 3)
		wg.Add(3)
		go func() {
			defer wg.Done()
			errs <- uc.AcceptInvitation(ctx, req)
		}()
		<-started
		for i := 0; i < 2; i++ {
			go func() {
				defer wg.Done()
				errs <- uc.AcceptInvitation(ctx, req)
			}()
		}
		time.Sleep(20 * time.Millisecond)
		close(release)
		wg.Wait()
		close(errs)

		for err := range errs {
			require.NoError(t, err)
		}
		deps.Enforcer.AssertNumberOfCalls(t, "AddGroupingPolicy", 1)
		deps.InvitationRepo.AssertNumberOfCalls(t, "Delete", 1)
	})

	t.Run("update member with unchanged role and status skips Casbin and write churn", func(t *testing.T) {
		deps, uc := setupMemberTest(t)
		setupDefaultRoleMocks(deps)
		actorID := "owner-1"
		ctx := usecase.WithActorUserID(context.Background(), actorID)
		orgID := "org-1"
		userID := "user-1"
		org := &entity.Organization{ID: orgID, OwnerID: actorID}
		member := &entity.OrganizationMember{
			OrganizationID: orgID,
			UserID:         userID,
			RoleID:         "role:member",
			Status:         entity.MemberStatusActive,
		}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.MemberRepo.On("FindMemberForUpdate", ctx, orgID, userID).Return(member, nil)
		deps.MemberRepo.On("GetMemberRoleName", ctx, orgID, userID).Return("role:member", nil)
		deps.MemberRepo.On("FindMembers", ctx, orgID).Return([]*entity.OrganizationMember{member}, nil)

		res, err := uc.UpdateMember(ctx, orgID, userID, &model.UpdateMemberRequest{
			RoleID: "role:member",
			Status: entity.MemberStatusActive,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, "role:member", res.RoleID)

		deps.MemberRepo.AssertNotCalled(t, "UpdateMemberRole", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		deps.MemberRepo.AssertNotCalled(t, "UpdateMemberStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		deps.Enforcer.AssertNotCalled(t, "RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		deps.Enforcer.AssertNotCalled(t, "AddGroupingPolicy", mock.Anything, mock.Anything, mock.Anything)
		deps.OrgReader.AssertNotCalled(t, "InvalidateMembershipCache", mock.Anything, mock.Anything, mock.Anything)
	})
}
