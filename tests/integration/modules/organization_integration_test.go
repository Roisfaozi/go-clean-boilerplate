package modules

import (
	"context"
	"fmt"
	"testing"
	"time"

	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertTestOrg(db *sqlx.DB, org *orgEntity.Organization) error {
	now := time.Now().UnixMilli()
	if org.CreatedAt == 0 {
		org.CreatedAt = now
	}
	if org.UpdatedAt == 0 {
		org.UpdatedAt = now
	}
	_, err := db.Exec(`INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
		org.ID, org.Name, org.Slug, org.OwnerID, org.Status, org.CreatedAt, org.UpdatedAt)
	return err
}

func insertTestMember(db *sqlx.DB, member *orgEntity.OrganizationMember) error {
	if member.JoinedAt == 0 {
		member.JoinedAt = time.Now().UnixMilli()
	}
	_, err := db.Exec(`INSERT INTO organization_members (id, organization_id, user_id, role_id, status, joined_at) VALUES (?, ?, ?, ?, ?, ?)`,
		member.ID, member.OrganizationID, member.UserID, member.RoleID, member.Status, member.JoinedAt)
	return err
}

// TestOrganizationRepository_Create tests atomic creation of org + owner member
func TestOrganizationRepository_Create(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	// Create a test user first
	user := setup.CreateTestUser(t, env.DB, "testowner", "owner@test.com", "password123")

	org := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Test Organization",
		Slug:    "test-org-" + uuid.New().String()[:8],
		OwnerID: user.ID,
		Status:  orgEntity.OrgStatusActive,
	}

	// Create org (will use repository once implemented)
	err := insertTestOrg(env.SQLXDB, org)
	require.NoError(t, err, "Should create organization")

	// Create owner membership
	member := &orgEntity.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		RoleID:         "role:admin",
		Status:         orgEntity.MemberStatusActive,
	}
	err = insertTestMember(env.SQLXDB, member)
	require.NoError(t, err, "Should create owner membership")

	// Verify organization was created
	var foundOrg orgEntity.Organization
	err = env.SQLXDB.Get(&foundOrg, "SELECT id, name, slug, owner_id, status, created_at, updated_at, deleted_at FROM organizations WHERE id = ? LIMIT 1", org.ID)
	require.NoError(t, err)
	assert.Equal(t, org.Name, foundOrg.Name)

	// Verify owner membership was created
	var foundMember orgEntity.OrganizationMember
	err = env.SQLXDB.Get(&foundMember, "SELECT id, organization_id, user_id, role_id, status, joined_at FROM organization_members WHERE organization_id = ? AND user_id = ? LIMIT 1", org.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, orgEntity.MemberStatusActive, foundMember.Status)
}

// TestOrganizationRepository_SlugUniqueness tests that slugs must be unique
func TestOrganizationRepository_SlugUniqueness(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	user := setup.CreateTestUser(t, env.DB, "sluguser", "slug@test.com", "password123")

	slug := "unique-slug-" + uuid.New().String()[:8]

	org1 := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Org 1",
		Slug:    slug,
		OwnerID: user.ID,
		Status:  orgEntity.OrgStatusActive,
	}
	err := insertTestOrg(env.SQLXDB, org1)
	require.NoError(t, err, "First org should be created")

	// Try to create another org with the same slug
	org2 := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Org 2",
		Slug:    slug, // Same slug
		OwnerID: user.ID,
		Status:  orgEntity.OrgStatusActive,
	}
	err = insertTestOrg(env.SQLXDB, org2)
	assert.Error(t, err, "Should fail with duplicate slug")
}

// TestOrganizationRepository_DataIsolation tests that queries are scoped by org_id
func TestOrganizationRepository_DataIsolation(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	// Create two users
	userA := setup.CreateTestUser(t, env.DB, "usera", "usera@test.com", "password123")
	userB := setup.CreateTestUser(t, env.DB, "userb", "userb@test.com", "password123")

	// Create two organizations
	orgA := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Org A",
		Slug:    "org-a-" + uuid.New().String()[:8],
		OwnerID: userA.ID,
		Status:  orgEntity.OrgStatusActive,
	}
	err := insertTestOrg(env.SQLXDB, orgA)
	require.NoError(t, err)

	orgB := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Org B",
		Slug:    "org-b-" + uuid.New().String()[:8],
		OwnerID: userB.ID,
		Status:  orgEntity.OrgStatusActive,
	}
	err = insertTestOrg(env.SQLXDB, orgB)
	require.NoError(t, err)

	// Add userA as member of orgA only
	memberA := &orgEntity.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: orgA.ID,
		UserID:         userA.ID,
		RoleID:         "role:admin",
		Status:         orgEntity.MemberStatusActive,
	}
	err = insertTestMember(env.SQLXDB, memberA)
	require.NoError(t, err)

	// Add userB as member of orgB only
	memberB := &orgEntity.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: orgB.ID,
		UserID:         userB.ID,
		RoleID:         "role:admin",
		Status:         orgEntity.MemberStatusActive,
	}
	err = insertTestMember(env.SQLXDB, memberB)
	require.NoError(t, err)

	// Test: Query members of orgA should NOT include userB
	var membersOfA []*orgEntity.OrganizationMember
	err = env.SQLXDB.Select(&membersOfA, "SELECT id, organization_id, user_id, role_id, status, joined_at FROM organization_members WHERE organization_id = ?", orgA.ID)
	require.NoError(t, err)

	assert.Len(t, membersOfA, 1, "Org A should have exactly 1 member")
	assert.Equal(t, userA.ID, membersOfA[0].UserID, "Org A member should be userA")

	// Test: Query members of orgB should NOT include userA
	var membersOfB []*orgEntity.OrganizationMember
	err = env.SQLXDB.Select(&membersOfB, "SELECT id, organization_id, user_id, role_id, status, joined_at FROM organization_members WHERE organization_id = ?", orgB.ID)
	require.NoError(t, err)

	assert.Len(t, membersOfB, 1, "Org B should have exactly 1 member")
	assert.Equal(t, userB.ID, membersOfB[0].UserID, "Org B member should be userB")
}

// TestOrganizationRepository_CheckMembership tests membership validation
func TestOrganizationRepository_CheckMembership(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	user := setup.CreateTestUser(t, env.DB, "memberuser", "member@test.com", "password123")

	org := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Member Test Org",
		Slug:    "member-test-" + uuid.New().String()[:8],
		OwnerID: user.ID,
		Status:  orgEntity.OrgStatusActive,
	}
	err := insertTestOrg(env.SQLXDB, org)
	require.NoError(t, err)

	// Add as active member
	member := &orgEntity.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		RoleID:         "role:user",
		Status:         orgEntity.MemberStatusActive,
	}
	err = insertTestMember(env.SQLXDB, member)
	require.NoError(t, err)

	// Test: Active member should be found
	var count int64
	err = env.SQLXDB.Get(&count, "SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ? AND status = ?", org.ID, user.ID, orgEntity.MemberStatusActive)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "Active member should be found")

	// Test: Non-existent member should not be found
	err = env.SQLXDB.Get(&count, "SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ? AND status = ?", org.ID, "non-existent-user", orgEntity.MemberStatusActive)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "Non-existent member should not be found")
}

// TestOrganizationRepository_BannedMemberDenied tests that banned members are rejected
func TestOrganizationRepository_BannedMemberDenied(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	user := setup.CreateTestUser(t, env.DB, "banneduser", "banned@test.com", "password123")

	org := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Ban Test Org",
		Slug:    "ban-test-" + uuid.New().String()[:8],
		OwnerID: user.ID,
		Status:  orgEntity.OrgStatusActive,
	}
	err := insertTestOrg(env.SQLXDB, org)
	require.NoError(t, err)

	// Add as banned member
	member := &orgEntity.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		RoleID:         "role:user",
		Status:         orgEntity.MemberStatusBanned,
	}
	err = insertTestMember(env.SQLXDB, member)
	require.NoError(t, err)

	// Test: Banned member should NOT be counted as active
	var count int64
	err = env.SQLXDB.Get(&count, "SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ? AND status = ?", org.ID, user.ID, orgEntity.MemberStatusActive)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "Banned member should not be counted as active")

	// Test: Banned member should still exist in table
	err = env.SQLXDB.Get(&count, "SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ?", org.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "Banned member record should exist")
}

// TestOrganizationRepository_AtomicCreate tests the repository's atomic create operation
func TestOrganizationRepository_AtomicCreate(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	orgRepo := orgRepo.NewOrganizationRepository(env.DB)

	user := setup.CreateTestUser(t, env.DB, "atomicowner", "atomic@test.com", "password123")

	org := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Atomic Test Org",
		Slug:    "atomic-test-" + uuid.New().String()[:8],
		OwnerID: user.ID,
		Status:  orgEntity.OrgStatusActive,
	}

	// Use repository's atomic create (creates org + owner member in transaction)
	err := orgRepo.Create(context.Background(), org, "role:admin")
	require.NoError(t, err, "Atomic create should succeed")

	// Verify organization was created
	foundOrg, err := orgRepo.FindByID(context.Background(), org.ID)
	require.NoError(t, err)
	require.NotNil(t, foundOrg)
	assert.Equal(t, org.Name, foundOrg.Name)

	// Verify owner membership was created automatically
	var count int64
	err = env.SQLXDB.Get(&count, "SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ? AND status = ?", org.ID, user.ID, orgEntity.MemberStatusActive)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "Owner membership should be created atomically")
}

// TestOrganizationMemberRepository_CheckMembership tests the member repository
func TestOrganizationMemberRepository_CheckMembership(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	memberRepo := orgRepo.NewOrganizationMemberRepository(env.DB)

	user := setup.CreateTestUser(t, env.DB, "membertest", "membertest@test.com", "password123")

	org := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Member Check Org",
		Slug:    "membercheck-" + uuid.New().String()[:8],
		OwnerID: user.ID,
		Status:  orgEntity.OrgStatusActive,
	}
	err := insertTestOrg(env.SQLXDB, org)
	require.NoError(t, err)

	// Add member
	member := &orgEntity.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		RoleID:         "role:member",
		Status:         orgEntity.MemberStatusActive,
	}
	err = memberRepo.AddMember(context.Background(), member)
	require.NoError(t, err)

	// Test: CheckMembership should return true for active member
	isMember, err := memberRepo.CheckMembership(context.Background(), org.ID, user.ID)
	require.NoError(t, err)
	assert.True(t, isMember, "Active member should be found")

	// Test: CheckMembership should return false for non-existent member
	isMember, err = memberRepo.CheckMembership(context.Background(), org.ID, "non-existent-user")
	require.NoError(t, err)
	assert.False(t, isMember, "Non-existent member should not be found")
}

func TestOrganizationMemberRepository_HardDeleteAllowsReAdd(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	memberRepo := orgRepo.NewOrganizationMemberRepository(env.DB)
	user := setup.CreateTestUser(t, env.DB, "harddelete", "harddelete@test.com", "password123")
	org := &orgEntity.Organization{
		ID:      uuid.New().String(),
		Name:    "Hard Delete Org",
		Slug:    "hard-delete-" + uuid.New().String()[:8],
		OwnerID: user.ID,
		Status:  orgEntity.OrgStatusActive,
	}
	require.NoError(t, insertTestOrg(env.SQLXDB, org))

	member := &orgEntity.OrganizationMember{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		RoleID:         "role:user",
		Status:         orgEntity.MemberStatusActive,
	}
	require.NoError(t, memberRepo.AddMember(context.Background(), member))
	require.NoError(t, memberRepo.RemoveMember(context.Background(), org.ID, user.ID))

	var count int64
	require.NoError(t, env.SQLXDB.Get(&count, "SELECT COUNT(*) FROM organization_members WHERE organization_id = ? AND user_id = ?", org.ID, user.ID))
	assert.Zero(t, count)

	member.ID = uuid.New().String()
	require.NoError(t, memberRepo.AddMember(context.Background(), member))
}

// TestOrganizationRepository_FindUserOrganizations tests finding orgs for a user
func TestOrganizationRepository_FindUserOrganizations(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	orgRepoImpl := orgRepo.NewOrganizationRepository(env.DB)

	user := setup.CreateTestUser(t, env.DB, "multiorguser", "multiorg@test.com", "password123")

	// Create 3 organizations and add user as member to 2 of them
	for i := 1; i <= 3; i++ {
		org := &orgEntity.Organization{
			ID:      uuid.New().String(),
			Name:    fmt.Sprintf("Org %d", i),
			Slug:    fmt.Sprintf("org-%d-%s", i, uuid.New().String()[:8]),
			OwnerID: user.ID,
			Status:  orgEntity.OrgStatusActive,
		}
		err := insertTestOrg(env.SQLXDB, org)
		require.NoError(t, err)

		// Only add user to first 2 orgs
		if i <= 2 {
			member := &orgEntity.OrganizationMember{
				ID:             uuid.New().String(),
				OrganizationID: org.ID,
				UserID:         user.ID,
				RoleID:         "role:member",
				Status:         orgEntity.MemberStatusActive,
			}
			err = insertTestMember(env.SQLXDB, member)
			require.NoError(t, err)
		}
	}

	// Test: FindUserOrganizations should return only the 2 orgs user is member of
	orgs, err := orgRepoImpl.FindUserOrganizations(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, orgs, 2, "User should be member of exactly 2 organizations")
}
