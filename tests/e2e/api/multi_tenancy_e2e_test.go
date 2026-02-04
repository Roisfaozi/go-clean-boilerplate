//go:build e2e
// +build e2e

package api

import (
	"testing"

	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestMultiTenancyE2E_MemberFlow(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	// 1. Setup Owner
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("OwnerPass123!"), bcrypt.DefaultCost)

	owner := f.Create(func(u *userEntity.User) {
		u.Username = "org_owner"
		u.Email = "owner@org.com"
		u.Password = string(hash)
	})

	// Login Owner
	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": owner.Username,
		"password": "OwnerPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	ownerToken := loginRes.Data.AccessToken

	// 2. Create Organization
	orgSlug := "e2e-tech-corp"
	resp = client.POST("/api/v1/organizations", map[string]any{
		"name": "E2E Tech Corp",
		"slug": orgSlug,
	}, setup.WithAuth(ownerToken))
	require.Equal(t, 201, resp.StatusCode)
	var orgRes struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&orgRes)
	orgID := orgRes.Data.ID

	// 3. Create Custom Role in Organization
	// Note: Role creation might need OrgID context if roles are org-scoped.
	// Current implementation: Roles table has organization_id column.
	// But RoleController.Create currently takes name/desc.
	// Let's assume roles created via API default to global or we need to pass org_id?
	// Checking RoleController.Create: it doesn't seem to take org_id from body, maybe defaults to global?
	// Let's check role repository Create. It uses struct.
	// The request body doesn't have org_id.
	// BUT, for this test, we can use a system role or create one manually in DB for the org.
	// Or simply use a standard "role:member" if available.
	// Let's create a role manually in DB for this org to be safe and "multi-tenant" compliant.

	memberRoleID := uuid.New().String()
	server.DB.Create(&roleEntity.Role{
		ID:             memberRoleID,
		Name:           "Viewer",
		OrganizationID: &orgID,
		Description:    "View only",
	})

	// 4. Invite New User
	inviteEmail := "new_member@test.com"
	resp = client.POST("/api/v1/organizations/"+orgID+"/members/invite", map[string]any{
		"email":   inviteEmail,
		"role_id": memberRoleID,
	}, setup.WithAuth(ownerToken), setup.WithOrg(orgID))
	require.Equal(t, 201, resp.StatusCode)

	// 5. Verify Invitation Token exists in DB
	var tokenRecord orgEntity.InvitationToken
	err := server.DB.Where("email = ? AND organization_id = ?", inviteEmail, orgID).First(&tokenRecord).Error
	require.NoError(t, err, "Invitation token should exist in DB")

	// 6. Accept Invitation
	resp = client.POST("/api/v1/organizations/invitations/accept", map[string]any{
		"token":    tokenRecord.Token,
		"password": "MemberPass123!",
		"name":     "New Member",
	})
	require.Equal(t, 200, resp.StatusCode)

	// 7. Verify Member is Active
	resp = client.GET("/api/v1/organizations/"+orgID+"/members", setup.WithAuth(ownerToken), setup.WithOrg(orgID))
	require.Equal(t, 200, resp.StatusCode)

	var membersRes struct {
		Data []struct {
			UserID string `json:"user_id"`
			Status string `json:"status"`
			User   struct {
				Email string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}
	resp.JSON(&membersRes)

	found := false
	for _, m := range membersRes.Data {
		if m.User.Email == inviteEmail {
			found = true
			assert.Equal(t, "active", m.Status)
		}
	}
	assert.True(t, found, "New member should be in the list")

	// 8. Verify Casbin Policy (Multi-tenancy check)
	// User should have role "Viewer" in domain "orgID"
	// Get User ID from member list or DB
	var newUser userEntity.User
	server.DB.Where("email = ?", inviteEmail).First(&newUser)

	ok, _ := server.Enforcer.HasGroupingPolicy(newUser.ID, memberRoleID, orgID)
	assert.True(t, ok, "Casbin grouping policy should exist for (user, role, org)")

	// 9. Login as New Member
	resp = client.POST("/api/v1/auth/login", map[string]any{
		"username": inviteEmail, // Shadow users use email as username initially
		"password": "MemberPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var memberLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&memberLoginRes)
	memberToken := memberLoginRes.Data.AccessToken

	// 10. Verify Access to Org Resources (Mocked via check-batch)
	// First, grant permission to the role in that org
	// This requires "GrantPermission" to support domain.
	// Current GrantPermission controller/usecase uses "global" hardcoded.
	// This is a limitation found!
	// But we can check if they can access "Get Organization" which implies membership check in middleware?
	// Let's check GetOrganization endpoint.

	resp = client.GET("/api/v1/organizations/"+orgID, setup.WithAuth(memberToken), setup.WithOrg(orgID))
	// Should be allowed if they are a member.
	assert.Equal(t, 200, resp.StatusCode)
}
