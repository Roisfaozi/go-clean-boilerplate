//go:build integration
// +build integration

package scenarios

import (
	"context"
	"errors"
	"testing"
	"time"

	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	orgUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	permissionUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// failOnAddEnforcer delegates every operation to the wrapped enforcer, but fails
// AddGroupingPolicy during a transaction to simulate a Casbin write error.
type failOnAddEnforcer struct {
	permissionUC.IEnforcer
}

func (e *failOnAddEnforcer) WithContext(ctx context.Context) permissionUC.IEnforcer {
	return &txBoundFailOnAdd{IEnforcer: e.IEnforcer.WithContext(ctx)}
}

type txBoundFailOnAdd struct {
	permissionUC.IEnforcer
}

func (e *txBoundFailOnAdd) WithContext(ctx context.Context) permissionUC.IEnforcer {
	return e
}

func (e *txBoundFailOnAdd) AddGroupingPolicy(params ...interface{}) (bool, error) {
	return false, errors.New("simulated casbin failure")
}

type memberFailureFixture struct {
	memberUC   orgUsecase.OrganizationMemberUseCase
	mRepo      orgRepo.OrganizationMemberRepository
	uRepo      userRepo.UserRepository
	orgID      string
	ownerID    string
	role1      *roleEntity.Role
	role2      *roleEntity.Role
	env        *setup.TestEnvironment
	actorCtx   context.Context
	enforcerUC permissionUC.IEnforcer
}

func newMemberFailureFixture(t *testing.T, env *setup.TestEnvironment, prefix string, enforcer permissionUC.IEnforcer) *memberFailureFixture {
	t.Helper()

	tm := tx.NewSQLXTransactionManager(env.SQLXDB, env.Logger)
	rRepo := roleRepo.NewRoleRepository(env.SQLXDB, env.Logger)
	mRepo := orgRepo.NewOrganizationMemberRepository(env.DB)
	oRepo := orgRepo.NewOrganizationRepository(env.DB, env.Redis)
	iRepo := orgRepo.NewInvitationRepository(env.DB)
	uRepo := userRepo.NewUserRepository(env.SQLXDB, env.Logger)
	oReader := orgUsecase.NewCachedOrgReader(mRepo, env.Redis, env.Logger)

	owner := setup.CreateTestUser(t, env.DB, prefix+"Owner", prefix+"owner@example.com", "Password123!")

	orgUC := orgUsecase.NewOrganizationUseCase(env.Logger, tm, oRepo, mRepo, oReader, env.Enforcer, rRepo)
	orgResp, err := orgUC.CreateOrganization(context.Background(), owner.ID, &orgModel.CreateOrganizationRequest{
		Name: prefix + " Org",
		Slug: prefix + "-org",
	})
	require.NoError(t, err)

	role1 := &roleEntity.Role{ID: uuid.NewString(), Name: prefix + "_role_1", OrganizationID: &orgResp.ID}
	role2 := &roleEntity.Role{ID: uuid.NewString(), Name: prefix + "_role_2", OrganizationID: &orgResp.ID}
	require.NoError(t, rRepo.Create(context.Background(), role1))
	require.NoError(t, rRepo.Create(context.Background(), role2))

	memberUC := orgUsecase.NewOrganizationMemberUseCase(
		env.Logger, tm, mRepo, oRepo, iRepo, uRepo, nil, enforcer, nil, oReader, "http://localhost:3000", rRepo,
	)

	actorCtx := orgUsecase.WithActorUserID(context.Background(), owner.ID)
	actorCtx = database.SetOrganizationContext(actorCtx, orgResp.ID)

	return &memberFailureFixture{
		memberUC:   memberUC,
		mRepo:      mRepo,
		uRepo:      uRepo,
		orgID:      orgResp.ID,
		ownerID:    owner.ID,
		role1:      role1,
		role2:      role2,
		env:        env,
		actorCtx:   actorCtx,
		enforcerUC: enforcer,
	}
}

func (f *memberFailureFixture) countGroupingPolicies(t *testing.T, userID string, roleName string) int64 {
	t.Helper()

	query := f.env.DB.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v2 = ?", "g", userID, f.orgID)
	if roleName != "" {
		query = query.Where("v1 = ?", roleName)
	}

	var count int64
	require.NoError(t, query.Count(&count).Error)
	return count
}

// TestUpdateMember_CasbinFailure_RollsBackDB verifies that a Casbin grouping
// policy failure rolls back the member role change in the database.
func TestUpdateMember_CasbinFailure_RollsBackDB(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	f := newMemberFailureFixture(t, env, "rollback", &failOnAddEnforcer{IEnforcer: env.Enforcer})

	member := setup.CreateTestUser(t, env.DB, "rollbackMember", "rollbackmember@example.com", "Password123!")
	require.NoError(t, f.mRepo.AddMember(context.Background(), &orgEntity.OrganizationMember{
		ID:             uuid.NewString(),
		OrganizationID: f.orgID,
		UserID:         member.ID,
		RoleID:         f.role1.ID,
		Status:         orgEntity.MemberStatusActive,
	}))
	_, err := env.Enforcer.AddGroupingPolicy(member.ID, f.role1.Name, f.orgID)
	require.NoError(t, err)

	tests := []struct {
		name              string
		request           *orgModel.UpdateMemberRequest
		expectError       bool
		expectedRoleID    string
		expectedPolicy    string
		expectedPolicyNum int64
	}{
		{
			name:              "casbin add failure rolls back role change and keeps old policy",
			request:           &orgModel.UpdateMemberRequest{RoleID: f.role2.ID},
			expectError:       true,
			expectedRoleID:    f.role1.ID,
			expectedPolicy:    f.role1.Name,
			expectedPolicyNum: 1,
		},
		{
			name:              "new role policy is never persisted after rollback",
			request:           &orgModel.UpdateMemberRequest{RoleID: f.role2.ID},
			expectError:       true,
			expectedRoleID:    f.role1.ID,
			expectedPolicy:    f.role2.Name,
			expectedPolicyNum: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.memberUC.UpdateMember(f.actorCtx, f.orgID, member.ID, tt.request)

			if tt.expectError {
				require.Error(t, err, "update must fail when Casbin add grouping policy fails")
			} else {
				require.NoError(t, err)
			}

			roleInDB, err := f.mRepo.GetMemberRole(context.Background(), f.orgID, member.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRoleID, roleInDB, "DB role must roll back when Casbin fails")

			assert.Equal(t, tt.expectedPolicyNum, f.countGroupingPolicies(t, member.ID, tt.expectedPolicy))
		})
	}
}

// TestAcceptInvitation_ActiveMember_NoDuplicatePolicy verifies invitation accept
// behaviour for invited versus already-active members.
func TestAcceptInvitation_ActiveMember_NoDuplicatePolicy(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	f := newMemberFailureFixture(t, env, "staleinvite", env.Enforcer)

	inviteEmail := "activemember@example.com"
	_, err := f.memberUC.InviteMember(f.actorCtx, f.orgID, &orgModel.InviteMemberRequest{
		Email:  inviteEmail,
		RoleID: f.role1.ID,
	})
	require.NoError(t, err)

	var firstToken orgEntity.InvitationToken
	require.NoError(t, env.DB.Where("organization_id = ? AND email = ?", f.orgID, inviteEmail).First(&firstToken).Error)

	staleToken := "stale_token_" + uuid.NewString()

	tests := []struct {
		name              string
		seedStaleToken    bool
		request           *orgModel.AcceptInvitationRequest
		expectedRoleID    string
		expectedPolicyNum int64
		consumedToken     string
	}{
		{
			name: "first accept activates member with a single policy",
			request: &orgModel.AcceptInvitationRequest{
				Token:    firstToken.Token,
				Password: "Password123!",
				Name:     "Active Member",
			},
			expectedRoleID:    f.role1.ID,
			expectedPolicyNum: 1,
			consumedToken:     firstToken.Token,
		},
		{
			name:              "stale accept on active member adds no duplicate policy",
			seedStaleToken:    true,
			request:           &orgModel.AcceptInvitationRequest{Token: staleToken},
			expectedRoleID:    f.role1.ID,
			expectedPolicyNum: 1,
			consumedToken:     staleToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.seedStaleToken {
				require.NoError(t, env.DB.Create(&orgEntity.InvitationToken{
					ID:             uuid.NewString(),
					OrganizationID: f.orgID,
					Email:          inviteEmail,
					Token:          staleToken,
					RoleID:         f.role2.ID,
					ExpiresAt:      time.Now().Add(48 * time.Hour).UnixMilli(),
					CreatedAt:      time.Now().UnixMilli(),
				}).Error)
			}

			require.NoError(t, f.memberUC.AcceptInvitation(context.Background(), tt.request))

			user, err := f.uRepo.FindByEmail(context.Background(), inviteEmail)
			require.NoError(t, err)

			roleInDB, err := f.mRepo.GetMemberRole(context.Background(), f.orgID, user.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRoleID, roleInDB, "member DB role must not change on stale accept")

			assert.Equal(t, tt.expectedPolicyNum, f.countGroupingPolicies(t, user.ID, ""),
				"active member must not receive a duplicate grouping policy")

			var tokenCount int64
			require.NoError(t, env.DB.Table("invitation_tokens").Where("token = ?", tt.consumedToken).Count(&tokenCount).Error)
			assert.Equal(t, int64(0), tokenCount, "consumed token must be cleaned up")
		})
	}
}
