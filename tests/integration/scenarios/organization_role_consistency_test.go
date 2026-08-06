//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	orgUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
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

// TestOrganizationRoleConsistency_Lifecycle verifies full integration lifecycle:
// DB stores roles.id, Casbin stores roles.name, update replaces policy with role.Name
func TestOrganizationRoleConsistency_Lifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	mRepo := orgRepo.NewOrganizationMemberRepository(env.DB)
	oRepo := orgRepo.NewOrganizationRepository(env.DB, env.Redis)
	iRepo := orgRepo.NewInvitationRepository(env.DB)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	oReader := orgUsecase.NewCachedOrgReader(mRepo, env.Redis, env.Logger)

	ownerUser := setup.CreateTestUser(t, env.DB, "orgowner", "orgowner@example.com", "Password123!")

	orgUC := orgUsecase.NewOrganizationUseCase(
		env.Logger, tm, oRepo, mRepo, oReader, env.Enforcer, rRepo,
	)

	createOrgReq := &orgModel.CreateOrganizationRequest{
		Name: "Role Consistency Org",
		Slug: "role-consistency-org",
	}
	orgResp, err := orgUC.CreateOrganization(context.Background(), ownerUser.ID, createOrgReq)
	require.NoError(t, err)
	require.NotNil(t, orgResp)
	orgID := orgResp.ID

	// Create custom role in org
	roleUUID := uuid.New().String()
	roleName := "custom_lead"
	customRole := &roleEntity.Role{
		ID:             roleUUID,
		Name:           roleName,
		OrganizationID: &orgID,
		Description:    "Custom Lead Role",
	}
	err = rRepo.Create(context.Background(), customRole)
	require.NoError(t, err)

	memberUC := orgUsecase.NewOrganizationMemberUseCase(
		env.Logger, tm, mRepo, oRepo, iRepo, uRepo, nil, env.Enforcer, nil, oReader, "http://localhost:3000", rRepo,
	)

	// Invite user with roleUUID (request.RoleID = roleUUID)
	inviteCtx := orgUsecase.WithActorUserID(context.Background(), ownerUser.ID)
	inviteCtx = database.SetOrganizationContext(inviteCtx, orgID)
	inviteReq := &orgModel.InviteMemberRequest{
		Email:  "member_custom@example.com",
		RoleID: roleUUID,
	}

	memberResp, err := memberUC.InviteMember(inviteCtx, orgID, inviteReq)
	require.NoError(t, err)
	require.NotNil(t, memberResp)
	assert.Equal(t, roleUUID, memberResp.RoleID)

	// Verify InvitationToken in DB stores role_id = roleUUID
	var invToken orgEntity.InvitationToken
	err = env.DB.Where("organization_id = ? AND email = ?", orgID, inviteReq.Email).First(&invToken).Error
	require.NoError(t, err)
	assert.Equal(t, roleUUID, invToken.RoleID)

	// Accept invitation
	acceptReq := &orgModel.AcceptInvitationRequest{
		Token:    invToken.Token,
		Password: "Password123!",
		Name:     "Custom Member",
	}
	err = memberUC.AcceptInvitation(context.Background(), acceptReq)
	require.NoError(t, err)

	// Verify DB stores role_id = roleUUID
	invitedUser, err := uRepo.FindByEmail(context.Background(), inviteReq.Email)
	require.NoError(t, err)

	memberRoleInDB, err := mRepo.GetMemberRole(context.Background(), orgID, invitedUser.ID)
	require.NoError(t, err)
	assert.Equal(t, roleUUID, memberRoleInDB)

	// Verify Casbin stores role.Name ("custom_lead"), NOT roleUUID
	var casbinCount int64
	err = env.DB.Table("casbin_rule").
		Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "g", invitedUser.ID, roleName, orgID).
		Count(&casbinCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), casbinCount)

	// Create second role and update member
	roleUUID2 := uuid.New().String()
	roleName2 := "custom_manager"
	customRole2 := &roleEntity.Role{
		ID:             roleUUID2,
		Name:           roleName2,
		OrganizationID: &orgID,
		Description:    "Custom Manager Role",
	}
	err = rRepo.Create(context.Background(), customRole2)
	require.NoError(t, err)

	updateReq := &orgModel.UpdateMemberRequest{
		RoleID: roleUUID2,
	}
	_, err = memberUC.UpdateMember(inviteCtx, orgID, invitedUser.ID, updateReq)
	require.NoError(t, err)

	// Verify DB updated to roleUUID2 and Casbin updated to roleName2
	memberRoleInDB2, err := mRepo.GetMemberRole(context.Background(), orgID, invitedUser.ID)
	require.NoError(t, err)
	assert.Equal(t, roleUUID2, memberRoleInDB2)

	err = env.DB.Table("casbin_rule").
		Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "g", invitedUser.ID, roleName2, orgID).
		Count(&casbinCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), casbinCount)

	// Remove member
	err = memberUC.RemoveMember(inviteCtx, orgID, invitedUser.ID)
	require.NoError(t, err)

	// Verify Casbin grouping rule removed
	err = env.DB.Table("casbin_rule").
		Where("ptype = ? AND v0 = ? AND v2 = ?", "g", invitedUser.ID, orgID).
		Count(&casbinCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), casbinCount)
}

func TestOrganizationRoleConsistency_ReinviteAndUpdateSuspended(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	mRepo := orgRepo.NewOrganizationMemberRepository(env.DB)
	oRepo := orgRepo.NewOrganizationRepository(env.DB, env.Redis)
	iRepo := orgRepo.NewInvitationRepository(env.DB)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	oReader := orgUsecase.NewCachedOrgReader(mRepo, env.Redis, env.Logger)

	ownerUser := setup.CreateTestUser(t, env.DB, "orgowner2", "orgowner2@example.com", "Password123!")

	orgUC := orgUsecase.NewOrganizationUseCase(
		env.Logger, tm, oRepo, mRepo, oReader, env.Enforcer, rRepo,
	)

	orgResp, err := orgUC.CreateOrganization(context.Background(), ownerUser.ID, &orgModel.CreateOrganizationRequest{
		Name: "Role Reinvite Org",
		Slug: "role-reinvite-org",
	})
	require.NoError(t, err)
	orgID := orgResp.ID

	// Create 2 custom roles
	role1ID := uuid.New().String()
	role1 := &roleEntity.Role{ID: role1ID, Name: "role_1", OrganizationID: &orgID}
	require.NoError(t, rRepo.Create(context.Background(), role1))

	role2ID := uuid.New().String()
	role2 := &roleEntity.Role{ID: role2ID, Name: "role_2", OrganizationID: &orgID}
	require.NoError(t, rRepo.Create(context.Background(), role2))

	memberUC := orgUsecase.NewOrganizationMemberUseCase(
		env.Logger, tm, mRepo, oRepo, iRepo, uRepo, nil, env.Enforcer, nil, oReader, "http://localhost:3000", rRepo,
	)

	inviteCtx := orgUsecase.WithActorUserID(context.Background(), ownerUser.ID)
	inviteCtx = database.SetOrganizationContext(inviteCtx, orgID)

	// 1. Initial Invite
	inviteEmail := "reinvite@example.com"
	_, err = memberUC.InviteMember(inviteCtx, orgID, &orgModel.InviteMemberRequest{
		Email:  inviteEmail,
		RoleID: role1ID,
	})
	require.NoError(t, err)

	// 2. Re-invite with role2 before acceptance (pending reinvite)
	_, err = memberUC.InviteMember(inviteCtx, orgID, &orgModel.InviteMemberRequest{
		Email:  inviteEmail,
		RoleID: role2ID,
	})
	require.NoError(t, err)

	// Verify InvitationToken in DB stores role_id = role2ID
	var invToken orgEntity.InvitationToken
	err = env.DB.Where("organization_id = ? AND email = ?", orgID, inviteEmail).First(&invToken).Error
	require.NoError(t, err)
	assert.Equal(t, role2ID, invToken.RoleID)

	// Accept invitation
	require.NoError(t, memberUC.AcceptInvitation(context.Background(), &orgModel.AcceptInvitationRequest{
		Token:    invToken.Token,
		Password: "Password123!",
	}))

	user, err := uRepo.FindByEmail(context.Background(), inviteEmail)
	require.NoError(t, err)

	// 3. Suspend member
	require.NoError(t, mRepo.UpdateMemberStatus(context.Background(), orgID, user.ID, orgEntity.MemberStatusSuspended))

	// 4. Remove suspended member -> should remove Casbin policy and member row without error
	require.NoError(t, memberUC.RemoveMember(inviteCtx, orgID, user.ID))

	var casbinCount int64
	err = env.DB.Table("casbin_rule").
		Where("ptype = ? AND v0 = ? AND v2 = ?", "g", user.ID, orgID).
		Count(&casbinCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), casbinCount)
}
