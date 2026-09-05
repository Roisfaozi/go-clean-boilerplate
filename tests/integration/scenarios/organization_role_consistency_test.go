//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	accessRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/repository"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	orgUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	permissionUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/usecase"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	roleRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	roleUseCase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/usecase"
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
	rRepo := roleRepo.NewRoleRepository(env.SQLXDB, env.Logger)
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

	err = env.DB.Table("casbin_rule").
		Where("ptype = ? AND v0 = ? AND v2 = ?", "g", invitedUser.ID, orgID).
		Count(&casbinCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), casbinCount)
}

func TestOrganizationAuth_AdminWithUUIDRoleID_CanManageMembers(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	rRepo := roleRepo.NewRoleRepository(env.SQLXDB, env.Logger)
	mRepo := orgRepo.NewOrganizationMemberRepository(env.DB)
	oRepo := orgRepo.NewOrganizationRepository(env.DB, env.Redis)
	iRepo := orgRepo.NewInvitationRepository(env.DB)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	oReader := orgUsecase.NewCachedOrgReader(mRepo, env.Redis, env.Logger)

	ownerUser := setup.CreateTestUser(t, env.DB, "orgowner3", "orgowner3@example.com", "Password123!")
	adminUser := setup.CreateTestUser(t, env.DB, "orgadmin3", "orgadmin3@example.com", "Password123!")

	orgUC := orgUsecase.NewOrganizationUseCase(
		env.Logger, tm, oRepo, mRepo, oReader, env.Enforcer, rRepo,
	)

	orgResp, err := orgUC.CreateOrganization(context.Background(), ownerUser.ID, &orgModel.CreateOrganizationRequest{
		Name: "UUID Admin Org",
		Slug: "uuid-admin-org",
	})
	require.NoError(t, err)
	orgID := orgResp.ID

	// Admin role ID is a UUID in production database, role name is "role:admin"
	adminRoleID := setup.RoleIDByName(t, env.DB, "role:admin")

	// Add adminUser as member with RoleID = adminRoleID (UUID)
	err = mRepo.AddMember(context.Background(), &orgEntity.OrganizationMember{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		UserID:         adminUser.ID,
		RoleID:         adminRoleID,
		Status:         orgEntity.MemberStatusActive,
	})
	require.NoError(t, err)
	_, err = env.Enforcer.AddGroupingPolicy(adminUser.ID, "role:admin", orgID)
	require.NoError(t, err)

	memberUC := orgUsecase.NewOrganizationMemberUseCase(
		env.Logger, tm, mRepo, oRepo, iRepo, uRepo, nil, env.Enforcer, nil, oReader, "http://localhost:3000", rRepo,
	)

	// Admin user with UUID role_id performs member management (InviteMember)
	adminCtx := orgUsecase.WithActorUserID(context.Background(), adminUser.ID)
	adminCtx = database.SetOrganizationContext(adminCtx, orgID)

	customRoleID := uuid.NewString()
	require.NoError(t, rRepo.Create(context.Background(), &roleEntity.Role{
		ID: customRoleID, Name: "custom_sub", OrganizationID: &orgID,
	}))

	_, err = memberUC.InviteMember(adminCtx, orgID, &orgModel.InviteMemberRequest{
		Email:  "sub_member@example.com",
		RoleID: customRoleID,
	})
	require.NoError(t, err, "Admin user with UUID role_id in DB must be authorized to manage members")
}

func TestDeleteRole_ScopedToOrganization(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	rRepo := roleRepo.NewRoleRepository(env.SQLXDB, env.Logger)
	mRepo := orgRepo.NewOrganizationMemberRepository(env.DB)
	oRepo := orgRepo.NewOrganizationRepository(env.DB, env.Redis)
	oReader := orgUsecase.NewCachedOrgReader(mRepo, env.Redis, env.Logger)

	ownerA := setup.CreateTestUser(t, env.DB, "ownerA", "ownerA@example.com", "Password123!")
	ownerB := setup.CreateTestUser(t, env.DB, "ownerB", "ownerB@example.com", "Password123!")

	orgUC := orgUsecase.NewOrganizationUseCase(env.Logger, tm, oRepo, mRepo, oReader, env.Enforcer, rRepo)
	roleUC := roleRepo.NewRoleRepository(env.SQLXDB, env.Logger)
	_ = roleUC

	orgA, _ := orgUC.CreateOrganization(context.Background(), ownerA.ID, &orgModel.CreateOrganizationRequest{Name: "Org A", Slug: "org-a"})
	orgB, _ := orgUC.CreateOrganization(context.Background(), ownerB.ID, &orgModel.CreateOrganizationRequest{Name: "Org B", Slug: "org-b"})

	// Both orgs create a role named "shared_name"
	roleAID := uuid.NewString()
	roleA := &roleEntity.Role{ID: roleAID, Name: "shared_name", OrganizationID: &orgA.ID}
	require.NoError(t, rRepo.Create(context.Background(), roleA))

	roleBID := uuid.NewString()
	roleB := &roleEntity.Role{ID: roleBID, Name: "shared_name", OrganizationID: &orgB.ID}
	require.NoError(t, rRepo.Create(context.Background(), roleB))

	userA := setup.CreateTestUser(t, env.DB, "userA", "userA@example.com", "Password123!")
	userB := setup.CreateTestUser(t, env.DB, "userB", "userB@example.com", "Password123!")

	env.Enforcer.AddGroupingPolicy(userA.ID, "shared_name", orgA.ID)
	env.Enforcer.AddGroupingPolicy(userB.ID, "shared_name", orgB.ID)

	// Delete custom role from Org A using DeleteForOrganization
	aRepo := accessRepo.NewAccessRepository(env.SQLXDB, env.Logger)
	usrRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	pUC := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, usrRepo, aRepo, nil)
	rUC := roleUseCase.NewRoleUseCase(env.Logger, tm, rRepo, pUC)
	require.NoError(t, rUC.DeleteForOrganization(context.Background(), orgA.ID, roleAID))

	var countA, countB int64
	env.DB.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "g", userA.ID, "shared_name", orgA.ID).Count(&countA)
	env.DB.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "g", userB.ID, "shared_name", orgB.ID).Count(&countB)

	assert.Equal(t, int64(0), countA, "Org A grouping policy should be deleted")
	assert.Equal(t, int64(1), countB, "Org B grouping policy must remain intact")
}

func TestOrganizationRoleConsistency_ReinviteAndUpdateSuspended(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	tm := tx.NewTransactionManager(env.DB, env.Logger)
	rRepo := roleRepo.NewRoleRepository(env.SQLXDB, env.Logger)
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
