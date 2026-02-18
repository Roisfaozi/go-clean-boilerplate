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
	t.Run("InviteMember - Error Scenarios", func(t *testing.T) {
		testCases := []struct {
			name        string
			orgID       string
			req         *model.InviteMemberRequest
			mockSetup   func(deps *memberTestDeps, ctx context.Context)
			expectedErr error
		}{
			{
				name:  "Org Find Error",
				orgID: "org-1",
				req:   &model.InviteMemberRequest{Email: "test@example.com"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:  "Org Not Found",
				orgID: "org-1",
				req:   &model.InviteMemberRequest{Email: "test@example.com"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, nil)
				},
				expectedErr: exception.ErrNotFound,
			},
			{
				name:  "User Create Error (Shadow)",
				orgID: "org-1",
				req:   &model.InviteMemberRequest{Email: "new@example.com"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					org := &entity.Organization{ID: "org-1"}
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
					deps.UserRepo.On("FindByEmail", ctx, "new@example.com").Return(nil, nil)
					deps.UserRepo.On("Create", ctx, mock.Anything).Return(errors.New("create error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:  "Check Membership Error",
				orgID: "org-1",
				req:   &model.InviteMemberRequest{Email: "existing@example.com"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					org := &entity.Organization{ID: "org-1"}
					user := &userEntity.User{ID: "u1", Email: "existing@example.com"}
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
					deps.UserRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, errors.New("db error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:  "Get Member Status Error",
				orgID: "org-1",
				req:   &model.InviteMemberRequest{Email: "existing@example.com"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					org := &entity.Organization{ID: "org-1"}
					user := &userEntity.User{ID: "u1", Email: "existing@example.com"}
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
					deps.UserRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, nil)
					deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", user.ID).Return("", errors.New("db error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:  "Invitation Create Error",
				orgID: "org-1",
				req:   &model.InviteMemberRequest{Email: "existing@example.com"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					org := &entity.Organization{ID: "org-1"}
					user := &userEntity.User{ID: "u1", Email: "existing@example.com"}
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
					deps.UserRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, nil)
					deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", user.ID).Return("", nil)
					deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
					deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, user.Email, "org-1").Return(nil)
					deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:  "User Find Error",
				orgID: "org-1",
				req:   &model.InviteMemberRequest{Email: "error@example.com"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					org := &entity.Organization{ID: "org-1"}
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
					deps.UserRepo.On("FindByEmail", ctx, "error@example.com").Return(nil, errors.New("db error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:  "Delete Old Invitation Error",
				orgID: "org-1",
				req:   &model.InviteMemberRequest{Email: "test@example.com"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					org := &entity.Organization{ID: "org-1"}
					user := &userEntity.User{ID: "u1", Email: "test@example.com"}
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
					deps.UserRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, nil)
					deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", user.ID).Return("", nil)
					deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
					deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, user.Email, "org-1").Return(errors.New("db error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:  "Task Distribute Error (Should Log and Proceed)",
				orgID: "org-1",
				req:   &model.InviteMemberRequest{Email: "test@example.com"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					org := &entity.Organization{ID: "org-1"}
					user := &userEntity.User{ID: "u1", Email: "test@example.com"}
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
					deps.UserRepo.On("FindByEmail", ctx, user.Email).Return(user, nil)
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", user.ID).Return(false, nil)
					deps.MemberRepo.On("GetMemberStatus", ctx, "org-1", user.ID).Return("", nil)
					deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
					deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, user.Email, "org-1").Return(nil)
					deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(nil)
					deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).Return(errors.New("queue error"))
				},
				expectedErr: nil, // Success expected despite task error
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					tc.mockSetup(deps, ctx)
					_ = fn(ctx)
				}).Return(tc.expectedErr)

				res, err := uc.InviteMember(ctx, tc.orgID, tc.req)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, res)
				}
			})
		}
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

	t.Run("UpdateMember - Scenarios", func(t *testing.T) {
		testCases := []struct {
			name        string
			orgID       string
			userID      string
			req         *model.UpdateMemberRequest
			mockSetup   func(deps *memberTestDeps, ctx context.Context)
			expectedErr error
		}{
			{
				name:   "Enforcer Remove Error (Should Proceed)",
				orgID:  "org-1",
				userID: "u1",
				req:    &model.UpdateMemberRequest{RoleID: "new-role"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(true, nil)
					deps.MemberRepo.On("UpdateMemberRole", ctx, "org-1", "u1", "new-role").Return(nil)
					deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, "u1", "", "org-1").Return(false, errors.New("casbin error"))
					deps.Enforcer.On("AddGroupingPolicy", "u1", "new-role", "org-1").Return(true, nil)
					deps.MemberRepo.On("FindMembers", ctx, "org-1").Return([]*entity.OrganizationMember{{UserID: "u1", RoleID: "new-role"}}, nil)
				},
				expectedErr: nil,
			},
			{
				name:   "Enforcer Add Error",
				orgID:  "org-1",
				userID: "u1",
				req:    &model.UpdateMemberRequest{RoleID: "new-role"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(true, nil)
					deps.MemberRepo.On("UpdateMemberRole", ctx, "org-1", "u1", "new-role").Return(nil)
					deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, "u1", "", "org-1").Return(true, nil)
					deps.Enforcer.On("AddGroupingPolicy", "u1", "new-role", "org-1").Return(false, errors.New("casbin error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:   "Status Update Success",
				orgID:  "org-1",
				userID: "u1",
				req:    &model.UpdateMemberRequest{Status: "active"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(true, nil)
					deps.MemberRepo.On("UpdateMemberStatus", ctx, "org-1", "u1", "active").Return(nil)
					deps.MemberRepo.On("FindMembers", ctx, "org-1").Return([]*entity.OrganizationMember{{UserID: "u1", Status: "active"}}, nil)
				},
				expectedErr: nil,
			},
			{
				name:   "Status Update Error",
				orgID:  "org-1",
				userID: "u1",
				req:    &model.UpdateMemberRequest{Status: "active"},
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(true, nil)
					deps.MemberRepo.On("UpdateMemberStatus", ctx, "org-1", "u1", "active").Return(errors.New("db error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					tc.mockSetup(deps, ctx)
					_ = fn(ctx)
				}).Return(tc.expectedErr)

				res, err := uc.UpdateMember(ctx, tc.orgID, tc.userID, tc.req)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, res)
				}
			})
		}
	})

	t.Run("RemoveMember - Scenarios", func(t *testing.T) {
		testCases := []struct {
			name        string
			orgID       string
			userID      string
			mockSetup   func(deps *memberTestDeps, ctx context.Context)
			expectedErr error
		}{
			{
				name:   "Check Membership Error",
				orgID:  "org-1",
				userID: "u1",
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(false, errors.New("db error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:   "Org Find Error",
				orgID:  "org-1",
				userID: "u1",
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(true, nil)
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))
				},
				expectedErr: exception.ErrInternalServer,
			},
			{
				name:   "Enforcer Error (Should Log and Proceed)",
				orgID:  "org-1",
				userID: "u1",
				mockSetup: func(deps *memberTestDeps, ctx context.Context) {
					org := &entity.Organization{ID: "org-1", OwnerID: "other"}
					deps.MemberRepo.On("CheckMembership", ctx, "org-1", "u1").Return(true, nil)
					deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
					deps.Enforcer.On("RemoveFilteredGroupingPolicy", 0, "u1", "", "org-1").Return(false, errors.New("casbin error"))
					deps.MemberRepo.On("RemoveMember", ctx, "org-1", "u1").Return(nil)
				},
				expectedErr: nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					tc.mockSetup(deps, ctx)
					_ = fn(ctx)
				}).Return(tc.expectedErr)

				err := uc.RemoveMember(ctx, tc.orgID, tc.userID)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				} else {
					assert.NoError(t, err)
				}
			})
		}
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
