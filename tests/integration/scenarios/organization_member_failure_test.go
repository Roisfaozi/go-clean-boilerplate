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

// TestUpdateMember_CasbinFailure_RollsBackDB verifies that when the Casbin
// grouping policy update fails, the whole operation (including the DB role
// change) is rolled back.
func TestUpdateMember_CasbinFailure_RollsBackDB(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	mRepo := orgRepo.NewOrganizationMemberRepository(env.DB)
	oRepo := orgRepo.NewOrganizationRepository(env.DB, env.Redis)
	iRepo := orgRepo.NewInvitationRepository(env.DB)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	oReader := orgUsecase.NewCachedOrgReader(mRepo, env.Redis, env.Logger)

	owner := setup.CreateTestUser(t, env.DB, "rollbackOwner", "rollbackowner@example.com", "Password123!")

	orgUC := orgUsecase.NewOrganizationUseCase(env.Logger, tm, oRepo, mRepo, oReader, env.Enforcer, rRepo)
	orgResp, err := orgUC.CreateOrganization(context.Background(), owner.ID, &orgModel.CreateOrganizationRequest{
		Name: "Rollback Org",
		Slug: "rollback-org",
	})
	require.NoError(t, err)
	orgID := orgResp.ID

	role1 := &roleEntity.Role{ID: uuid.NewString(), Name: "role_1", OrganizationID: &orgID}
	role2 := &roleEntity.Role{ID: uuid.NewString(), Name: "role_2", OrganizationID: &orgID}
	require.NoError(t, rRepo.Create(context.Background(), role1))
	require.NoError(t, rRepo.Create(context.Background(), role2))

	memberUser := setup.CreateTestUser(t, env.DB, "rollbackMember", "rollbackmember@example.com", "password")
	require.NoError(t, mRepo.AddMember(context.Background(), &orgEntity.OrganizationMember{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		UserID:         memberUser.ID,
		RoleID:         role1.ID,
		Status:         orgEntity.MemberStatusActive,
	}))
	_, err = env.Enforcer.AddGroupingPolicy(memberUser.ID, role1.Name, orgID)
	require.NoError(t, err)

	failing := &failOnAddEnforcer{IEnforcer: env.Enforcer}
	memberUC := orgUsecase.NewOrganizationMemberUseCase(
		env.Logger, tm, mRepo, oRepo, iRepo, uRepo, nil, failing, nil, oReader, "http://localhost:3000", rRepo,
	)

	actorCtx := orgUsecase.WithActorUserID(context.Background(), owner.ID)
	actorCtx = database.SetOrganizationContext(actorCtx, orgID)

	_, err = memberUC.UpdateMember(actorCtx, orgID, memberUser.ID, &orgModel.UpdateMemberRequest{RoleID: role2.ID})
	require.Error(t, err, "update must fail when Casbin add grouping policy fails")

	roleInDB, err := mRepo.GetMemberRole(context.Background(), orgID, memberUser.ID)
	require.NoError(t, err)
	assert.Equal(t, role1.ID, roleInDB, "DB role must roll back to old role when Casbin fails")

	var casbinCount int64
	err = env.DB.Table("casbin_rule").
		Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "g", memberUser.ID, role1.Name, orgID).
		Count(&casbinCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), casbinCount, "old grouping policy must remain after rollback")
}

// TestAcceptInvitation_ActiveMember_NoDuplicatePolicy verifies that accepting a
// stale/second invitation for an already-active member does not add a second
// Casbin grouping policy and cleans up the stale token.
func TestAcceptInvitation_ActiveMember_NoDuplicatePolicy(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	mRepo := orgRepo.NewOrganizationMemberRepository(env.DB)
	oRepo := orgRepo.NewOrganizationRepository(env.DB, env.Redis)
	iRepo := orgRepo.NewInvitationRepository(env.DB)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	oReader := orgUsecase.NewCachedOrgReader(mRepo, env.Redis, env.Logger)

	ownerUser := setup.CreateTestUser(t, env.DB, "activeOwner", "activeowner@example.com", "password")

	orgUC := orgUsecase.NewOrganizationUseCase(env.Logger, tm, oRepo, mRepo, oReader, env.Enforcer, rRepo)
	orgResp, err := orgUC.CreateOrganization(context.Background(), ownerUser.ID, &orgModel.CreateOrganizationRequest{
		Name: "Stale Invite Org",
		Slug: "stale-invite-org",
	})
	require.NoError(t, err)
	orgID := orgResp.ID

	role1 := &roleEntity.Role{ID: uuid.NewString(), Name: "role_1", OrganizationID: &orgID}
	role2 := &roleEntity.Role{ID: uuid.NewString(), Name: "role_2", OrganizationID: &orgID}
	require.NoError(t, rRepo.Create(context.Background(), role1))
	require.NoError(t, rRepo.Create(context.Background(), role2))

	memberUC := orgUsecase.NewOrganizationMemberUseCase(
		env.Logger, tm, mRepo, oRepo, iRepo, uRepo, nil, env.Enforcer, nil, oReader, "http://localhost:3000", rRepo,
	)

	inviteEmail := "activemember@example.com"
	inviteCtx := orgUsecase.WithActorUserID(context.Background(), ownerUser.ID)
	inviteCtx = database.SetOrganizationContext(inviteCtx, orgID)

	// 1. Invite with role1 and accept -> member becomes active with one policy
	_, err = memberUC.InviteMember(inviteCtx, orgID, &orgModel.InviteMemberRequest{Email: inviteEmail, RoleID: role1.ID})
	require.NoError(t, err)

	var firstToken orgEntity.InvitationToken
	require.NoError(t, env.DB.Where("organization_id = ? AND email = ?", orgID, inviteEmail).First(&firstToken).Error)
	require.NoError(t, memberUC.AcceptInvitation(context.Background(), &orgModel.AcceptInvitationRequest{
		Token:    firstToken.Token,
		Password: "Password123!",
		Name:     "Active Member",
	}))

	user, err := uRepo.FindByEmail(context.Background(), inviteEmail)
	require.NoError(t, err)

	roleInDB, err := mRepo.GetMemberRole(context.Background(), orgID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, role1.ID, roleInDB)

	// 2. Create a stale second invitation token with role2 for the same user
	staleToken := "stale_token_" + uuid.NewString()
	require.NoError(t, env.DB.Create(&orgEntity.InvitationToken{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		Email:          inviteEmail,
		Token:          staleToken,
		RoleID:         role2.ID,
		ExpiresAt:      time.Now().Add(48 * time.Hour).UnixMilli(),
		CreatedAt:      time.Now().UnixMilli(),
	}).Error)

	// 3. Accept stale token -> must succeed without adding a second policy
	require.NoError(t, memberUC.AcceptInvitation(context.Background(), &orgModel.AcceptInvitationRequest{Token: staleToken}))

	var casbinCount int64
	require.NoError(t, env.DB.Table("casbin_rule").
		Where("ptype = ? AND v0 = ? AND v2 = ?", "g", user.ID, orgID).
		Count(&casbinCount).Error)
	assert.Equal(t, int64(1), casbinCount, "active member must not get a second grouping policy")

	roleInDB, err = mRepo.GetMemberRole(context.Background(), orgID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, role1.ID, roleInDB, "member DB role must not change on stale token accept")

	var tokenCount int64
	require.NoError(t, env.DB.Table("invitation_tokens").Where("token = ?", staleToken).Count(&tokenCount).Error)
	assert.Equal(t, int64(0), tokenCount, "stale token must be cleaned up")
}
