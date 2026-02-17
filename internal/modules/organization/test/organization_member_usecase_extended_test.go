package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrganizationMemberUseCase_Extended(t *testing.T) {
	t.Run("InviteMember - Org Find Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.InviteMemberRequest{Email: "test@example.com"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))

		_, err := uc.InviteMember(ctx, "org-1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("InviteMember - Org Not Found", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.InviteMemberRequest{Email: "test@example.com"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrNotFound)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, nil)

		_, err := uc.InviteMember(ctx, "org-1", req)
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("InviteMember - User Create Error (Shadow)", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		org := &entity.Organization{ID: "org-1"}
		req := &model.InviteMemberRequest{Email: "new@example.com"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, nil) // Not found
		deps.UserRepo.On("Create", ctx, mock.Anything).Return(errors.New("create error"))

		_, err := uc.InviteMember(ctx, "org-1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("InviteMember - Check Membership Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		org := &entity.Organization{ID: "org-1"}
		req := &model.InviteMemberRequest{Email: "existing@example.com"}
		user := &userEntity.User{ID: "u1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, errors.New("db error"))

		_, err := uc.InviteMember(ctx, "org-1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("InviteMember - Get Member Status Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		org := &entity.Organization{ID: "org-1"}
		req := &model.InviteMemberRequest{Email: "existing@example.com"}
		user := &userEntity.User{ID: "u1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", user.ID).Return("", errors.New("db error"))

		_, err := uc.InviteMember(ctx, "org-1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("InviteMember - Invitation Create Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		org := &entity.Organization{ID: "org-1"}
		req := &model.InviteMemberRequest{Email: "existing@example.com"}
		user := &userEntity.User{ID: "u1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, "org-1").Return(nil)
		deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

		_, err := uc.InviteMember(ctx, "org-1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("AcceptInvitation - User Update Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
		inv := &entity.InvitationToken{ID: "inv-1", Email: "new@example.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "u1", Email: "new@example.com", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.UserRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("InviteMember - User Find Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.InviteMemberRequest{Email: "error@example.com"}
		org := &entity.Organization{ID: "org-1"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("db error"))

		_, err := uc.InviteMember(ctx, "org-1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("InviteMember - Delete Old Invitation Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.InviteMemberRequest{Email: "test@example.com"}
		org := &entity.Organization{ID: "org-1"}
		user := &userEntity.User{ID: "u1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, "org-1").Return(errors.New("db error"))

		_, err := uc.InviteMember(ctx, "org-1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("AcceptInvitation - User Find Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{ID: "inv-1", Email: "u@e.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(nil, errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("AcceptInvitation - Get Member Status Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{ID: "inv-1", OrganizationID: "org-1", Email: "u@e.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "u1", Email: "u@e.com", Status: "active"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", "u1").Return("", errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("AcceptInvitation - Enforcer Add Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{ID: "inv-1", OrganizationID: "org-1", Email: "u@e.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "u1", Email: "u@e.com", Status: "active"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", "u1").Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.Enforcer.On("AddGroupingPolicy", "u1", mock.Anything, "org-1").Return(false, errors.New("casbin error"))

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("AcceptInvitation - Missing Password for New User", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"} // No password
		inv := &entity.InvitationToken{ID: "inv-1", Email: "new@example.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "u1", Email: "new@example.com", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrBadRequest)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrBadRequest)
	})

	t.Run("UpdateMember - Enforcer Remove Error (Should Proceed)", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "u1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)

		// Fail remove, but proceed to add
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "", orgID).Return(false, errors.New("casbin error"))
		deps.Enforcer.On("AddGroupingPolicy", userID, "new-role", orgID).Return(true, nil)

		deps.MemberRepo.On("FindMembers", ctx, orgID).Return([]*entity.OrganizationMember{{UserID: userID, RoleID: "new-role"}}, nil)

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		assert.NoError(t, err)
	})

	t.Run("UpdateMember - Enforcer Add Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "u1"
		req := &model.UpdateMemberRequest{RoleID: "new-role"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)

		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "", orgID).Return(true, nil)
		deps.Enforcer.On("AddGroupingPolicy", userID, "new-role", orgID).Return(false, errors.New("casbin error"))

		_, err := uc.UpdateMember(ctx, orgID, userID, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("RemoveMember - Check Membership Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "u1"

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(false, errors.New("db error"))

		err := uc.RemoveMember(ctx, orgID, userID)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("InviteMember - Task Distribute Error (Should Log and Proceed)", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.InviteMemberRequest{Email: "test@example.com"}
		org := &entity.Organization{ID: "org-1"}
		user := &userEntity.User{ID: "u1", Email: req.Email}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
		deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
		deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, "org-1").Return(nil)
		deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(nil)

		// Task error
		deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).Return(errors.New("queue error"))

		res, err := uc.InviteMember(ctx, "org-1", req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("AcceptInvitation - Token Expired", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		// Expired
		inv := &entity.InvitationToken{ID: "inv-1", ExpiresAt: time.Now().Add(-1 * time.Hour).UnixMilli()}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrUnauthorized)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrUnauthorized)
	})

	t.Run("AcceptInvitation - User Not Found (Anomaly)", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{ID: "inv-1", Email: "missing@example.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrNotFound)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(nil, nil) // User missing

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrNotFound)
	})

	t.Run("AcceptInvitation - Update Member Status Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{ID: "inv-1", OrganizationID: "org-1", Email: "u@e.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "u1", Email: "u@e.com", Status: "active"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", "u1").Return(entity.MemberStatusInvited, nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, "org-1", "u1", entity.MemberStatusActive).Return(errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("AcceptInvitation - Add Member Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{ID: "inv-1", OrganizationID: "org-1", Email: "u@e.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "u1", Email: "u@e.com", Status: "active"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", "u1").Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("AcceptInvitation - Delete Invitation Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{ID: "inv-1", OrganizationID: "org-1", Email: "u@e.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "u1", Email: "u@e.com", Status: "active"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", "u1").Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.Enforcer.On("AddGroupingPolicy", "u1", mock.Anything, "org-1").Return(true, nil)
		deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(errors.New("db error"))

		err := uc.AcceptInvitation(ctx, req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("UpdateMember - Status Update Success", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.UpdateMemberRequest{Status: "active"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(true, nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, "org-1", "u1", "active").Return(nil)
		deps.MemberRepo.On("FindMembers", ctx, "org-1").Return([]*entity.OrganizationMember{{UserID: "u1", Status: "active"}}, nil)

		res, err := uc.UpdateMember(ctx, "org-1", "u1", req)
		assert.NoError(t, err)
		assert.Equal(t, "active", res.Status)
	})

	t.Run("UpdateMember - Status Update Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.UpdateMemberRequest{Status: "active"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(true, nil)
		deps.MemberRepo.On("UpdateMemberStatus", ctx, "org-1", "u1", "active").Return(errors.New("db error"))

		_, err := uc.UpdateMember(ctx, "org-1", "u1", req)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("RemoveMember - Org Find Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(exception.ErrInternalServer)

		deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(true, nil)
		deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))

		err := uc.RemoveMember(ctx, "org-1", "u1")
		assert.ErrorIs(t, err, exception.ErrInternalServer)
	})

	t.Run("AcceptInvitation - Default Name Logic", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"} // Empty Name
		inv := &entity.InvitationToken{ID: "inv-1", Email: "new@example.com", OrganizationID: "org-1", Role: "role:member", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "user-1", Email: "new@example.com", Status: "invited"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		// Expect update with Name = Email
		deps.UserRepo.On("Update", ctx, mock.MatchedBy(func(u *userEntity.User) bool {
			return u.Name == user.Email
		})).Return(nil)
		deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return("", nil)
		deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
		deps.Enforcer.On("AddGroupingPolicy", user.ID, inv.Role, inv.OrganizationID).Return(true, nil)
		deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(nil)

		err := uc.AcceptInvitation(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("AcceptInvitation - Idempotent (Already Active)", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		req := &model.AcceptInvitationRequest{Token: "token"}
		inv := &entity.InvitationToken{ID: "inv-1", OrganizationID: "org-1", Email: "u@e.com", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
		user := &userEntity.User{ID: "u1", Email: "u@e.com", Status: "active"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
		deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
		// Member status is already active, so switch case goes to default (break)
		deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", "u1").Return(entity.MemberStatusActive, nil)
		deps.Enforcer.On("AddGroupingPolicy", "u1", mock.Anything, "org-1").Return(true, nil)
		deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(nil)

		err := uc.AcceptInvitation(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("RemoveMember - Enforcer Error (Should Log and Proceed)", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		orgID := "org-1"
		userID := "u1"
		org := &entity.Organization{ID: orgID, OwnerID: "other"}

		deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			_ = fn(ctx)
		}).Return(nil)

		deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
		deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
		deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, userID, "", orgID).Return(false, errors.New("casbin error"))
		deps.MemberRepo.On("RemoveMember", ctx, orgID, userID).Return(nil)

		err := uc.RemoveMember(ctx, orgID, userID)
		assert.NoError(t, err)
	})

	t.Run("GetMembers - Repo Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		deps.MemberRepo.On("FindMembers", ctx, "org-1").Return(nil, errors.New("db error"))

		res, err := uc.GetMembers(ctx, "org-1")
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("GetPresence - Presence Error", func(t *testing.T) {
		deps, uc := setupMemberTest()
		ctx := context.Background()
		deps.Presence.On("GetOnlineUsers", ctx, "org-1").Return(nil, errors.New("redis error"))

		res, err := uc.GetPresence(ctx, "org-1")
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}
