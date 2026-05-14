//go:build e2e
// +build e2e

package api

import (
	"net/http"
	"testing"

	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	projectEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/entity"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestCrossTenantIsolation_ProjectAccess tests that a user from Org A
// cannot read, update, or delete a project belonging to Org B.
// This validates Row-Level Security (database.OrganizationScope) and
// Casbin multi-tenant authorization working together.
func TestCrossTenantIsolation_ProjectAccess(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	// ── Setup ──

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("TestPass123!"), bcrypt.DefaultCost)

	// Create User A (owner of Org A)
	userA := f.Create(func(u *userEntity.User) {
		u.Username = "user_a_owner"
		u.Email = "user_a@orgA.com"
		u.Password = string(hash)
	})

	// Create User B (owner of Org B)
	userB := f.Create(func(u *userEntity.User) {
		u.Username = "user_b_owner"
		u.Email = "user_b@orgB.com"
		u.Password = string(hash)
	})

	// Login User A
	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": userA.Username,
		"password": "TestPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var loginA struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginA)
	tokenA := loginA.Data.AccessToken

	// Login User B
	resp = client.POST("/api/v1/auth/login", map[string]any{
		"username": userB.Username,
		"password": "TestPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var loginB struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginB)
	tokenB := loginB.Data.AccessToken

	// Create Org A
	resp = client.POST("/api/v1/organizations", map[string]any{
		"name": "Org Alpha",
		"slug": "org-alpha",
	}, setup.WithAuth(tokenA))
	require.Equal(t, 201, resp.StatusCode)
	var orgARes struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&orgARes)
	orgAID := orgARes.Data.ID

	// Create Org B
	resp = client.POST("/api/v1/organizations", map[string]any{
		"name": "Org Beta",
		"slug": "org-beta",
	}, setup.WithAuth(tokenB))
	require.Equal(t, 201, resp.StatusCode)
	var orgBRes struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&orgBRes)
	orgBID := orgBRes.Data.ID

	// Grant Casbin policies for both users in their respective orgs
	roleA := uuid.New().String()
	server.DB.Create(&roleEntity.Role{
		ID:             roleA,
		Name:           "ProjectManager",
		OrganizationID: &orgAID,
		Description:    "Can manage projects",
	})

	roleB := uuid.New().String()
	server.DB.Create(&roleEntity.Role{
		ID:             roleB,
		Name:           "ProjectManager",
		OrganizationID: &orgBID,
		Description:    "Can manage projects",
	})

	// Assign roles in Casbin
	_, err := server.Enforcer.AddGroupingPolicy(userA.ID, roleA, orgAID)
	require.NoError(t, err)
	_, err = server.Enforcer.AddGroupingPolicy(userB.ID, roleB, orgBID)
	require.NoError(t, err)

	// Grant project CRUD policies for each role in their own org
	_, err = server.Enforcer.AddPolicy(roleA, orgAID, "/api/v1/projects", "*")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy(roleA, orgAID, "/api/v1/projects/:id", "*")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy(roleB, orgBID, "/api/v1/projects", "*")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy(roleB, orgBID, "/api/v1/projects/:id", "*")
	require.NoError(t, err)

	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	// Create a project in Org A (directly in DB for isolation test)
	projectAlphaID := uuid.New().String()
	server.DB.Create(&projectEntity.Project{
		ID:             projectAlphaID,
		Name:           "Alpha Secret Project",
		Domain:         "alpha-secret",
		OrganizationID: orgAID,
		Status:         "active",
	})

	// Create a project in Org B (directly in DB)
	projectBetaID := uuid.New().String()
	server.DB.Create(&projectEntity.Project{
		ID:             projectBetaID,
		Name:           "Beta Secret Project",
		Domain:         "beta-secret",
		OrganizationID: orgBID,
		Status:         "active",
	})

	// ── Test Cases ──

	t.Run("User A can read own project", func(t *testing.T) {
		resp := client.GET("/api/v1/projects/"+projectAlphaID,
			setup.WithAuth(tokenA),
			setup.WithOrg(orgAID),
		)
		// Should succeed - User A is in Org A and project belongs to Org A
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("User B cannot read Org A project", func(t *testing.T) {
		resp := client.GET("/api/v1/projects/"+projectAlphaID,
			setup.WithAuth(tokenB),
			setup.WithOrg(orgAID), // Attempting to access Org A
		)
		// Should be denied - User B is not a member of Org A
		// TenantMiddleware should return 403 (not a member)
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"Expected 403 or 404 but got %d: %s", resp.StatusCode, resp.String())
	})

	t.Run("User B cannot read Org A project via own org context", func(t *testing.T) {
		resp := client.GET("/api/v1/projects/"+projectAlphaID,
			setup.WithAuth(tokenB),
			setup.WithOrg(orgBID), // Using own org, but project ID belongs to Org A
		)
		// Should return 404 - OrganizationScope filters by org_id,
		// so the project won't be found in Org B's context
		assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusInternalServerError,
			"Expected 404 (record not found) but got %d: %s", resp.StatusCode, resp.String())
	})

	t.Run("User B cannot update Org A project", func(t *testing.T) {
		resp := client.PUT("/api/v1/projects/"+projectAlphaID, map[string]any{
			"name": "Hacked by Org B",
		},
			setup.WithAuth(tokenB),
			setup.WithOrg(orgAID),
		)
		// Should be denied at the tenant middleware level
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"Expected 403 or 404 but got %d: %s", resp.StatusCode, resp.String())
	})

	t.Run("User B cannot delete Org A project", func(t *testing.T) {
		resp := client.DELETE("/api/v1/projects/"+projectAlphaID,
			setup.WithAuth(tokenB),
			setup.WithOrg(orgAID),
		)
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"Expected 403 or 404 but got %d: %s", resp.StatusCode, resp.String())
	})

	t.Run("User A cannot read Org B project", func(t *testing.T) {
		resp := client.GET("/api/v1/projects/"+projectBetaID,
			setup.WithAuth(tokenA),
			setup.WithOrg(orgBID),
		)
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"Expected 403 or 404 but got %d: %s", resp.StatusCode, resp.String())
	})

	t.Run("User B can read own project", func(t *testing.T) {
		resp := client.GET("/api/v1/projects/"+projectBetaID,
			setup.WithAuth(tokenB),
			setup.WithOrg(orgBID),
		)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestCrossTenantIsolation_MemberList tests that listing members
// of Org A is not possible for a user who belongs only to Org B.
func TestCrossTenantIsolation_MemberList(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("TestPass123!"), bcrypt.DefaultCost)

	// Create owners
	userA := f.Create(func(u *userEntity.User) {
		u.Username = "iso_user_a"
		u.Email = "iso_a@test.com"
		u.Password = string(hash)
	})
	userB := f.Create(func(u *userEntity.User) {
		u.Username = "iso_user_b"
		u.Email = "iso_b@test.com"
		u.Password = string(hash)
	})

	// Login both
	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": userA.Username,
		"password": "TestPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var la struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&la)
	tokenA := la.Data.AccessToken

	resp = client.POST("/api/v1/auth/login", map[string]any{
		"username": userB.Username,
		"password": "TestPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var lb struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&lb)
	tokenB := lb.Data.AccessToken

	// Create Orgs
	resp = client.POST("/api/v1/organizations", map[string]any{
		"name": "Isolated Org A",
		"slug": "iso-org-a",
	}, setup.WithAuth(tokenA))
	require.Equal(t, 201, resp.StatusCode)
	var orgA struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&orgA)

	resp = client.POST("/api/v1/organizations", map[string]any{
		"name": "Isolated Org B",
		"slug": "iso-org-b",
	}, setup.WithAuth(tokenB))
	require.Equal(t, 201, resp.StatusCode)
	var orgB struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&orgB)

	// Grant Casbin policies for member listing
	roleA := uuid.New().String()
	server.DB.Create(&roleEntity.Role{
		ID:             roleA,
		Name:           "OrgAdmin",
		OrganizationID: &orgA.Data.ID,
	})
	_, _ = server.Enforcer.AddGroupingPolicy(userA.ID, roleA, orgA.Data.ID)
	_, _ = server.Enforcer.AddPolicy(roleA, orgA.Data.ID, "/api/v1/organizations/:id/members", "GET")
	_ = server.Enforcer.SavePolicy()

	t.Run("User B cannot list Org A members", func(t *testing.T) {
		resp := client.GET("/api/v1/organizations/"+orgA.Data.ID+"/members",
			setup.WithAuth(tokenB),
			setup.WithOrg(orgA.Data.ID),
		)
		// User B is not a member of Org A - should be denied
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"Expected 403 or 404 but got %d: %s", resp.StatusCode, resp.String())
	})

	t.Run("User A can list own org members", func(t *testing.T) {
		resp := client.GET("/api/v1/organizations/"+orgA.Data.ID+"/members",
			setup.WithAuth(tokenA),
			setup.WithOrg(orgA.Data.ID),
		)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestCrossTenantIsolation_OrgDetails tests that a user from Org B
// cannot view, update or delete Org A's details.
func TestCrossTenantIsolation_OrgDetails(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("TestPass123!"), bcrypt.DefaultCost)

	userA := f.Create(func(u *userEntity.User) {
		u.Username = "orgdetail_a"
		u.Email = "orgdetail_a@test.com"
		u.Password = string(hash)
	})
	userB := f.Create(func(u *userEntity.User) {
		u.Username = "orgdetail_b"
		u.Email = "orgdetail_b@test.com"
		u.Password = string(hash)
	})

	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": userA.Username,
		"password": "TestPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var la struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&la)
	tokenA := la.Data.AccessToken

	resp = client.POST("/api/v1/auth/login", map[string]any{
		"username": userB.Username,
		"password": "TestPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var lb struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&lb)
	tokenB := lb.Data.AccessToken

	// Create Org A
	resp = client.POST("/api/v1/organizations", map[string]any{
		"name": "Detail Org A",
		"slug": "detail-org-a",
	}, setup.WithAuth(tokenA))
	require.Equal(t, 201, resp.StatusCode)
	var orgA struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&orgA)

	// Grant Casbin policies for User A
	roleA := uuid.New().String()
	server.DB.Create(&roleEntity.Role{
		ID: roleA, Name: "OrgViewer", OrganizationID: &orgA.Data.ID,
	})
	_, _ = server.Enforcer.AddGroupingPolicy(userA.ID, roleA, orgA.Data.ID)
	_, _ = server.Enforcer.AddPolicy(roleA, orgA.Data.ID, "/api/v1/organizations/:id", "*")
	_ = server.Enforcer.SavePolicy()

	// Suppress unused variable warning
	_ = orgEntity.InvitationToken{}

	t.Run("User B cannot view Org A details", func(t *testing.T) {
		resp := client.GET("/api/v1/organizations/"+orgA.Data.ID,
			setup.WithAuth(tokenB),
			setup.WithOrg(orgA.Data.ID),
		)
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"Expected 403 or 404 but got %d", resp.StatusCode)
	})

	t.Run("User B cannot update Org A", func(t *testing.T) {
		resp := client.PUT("/api/v1/organizations/"+orgA.Data.ID, map[string]any{
			"name": "Hijacked by B",
		},
			setup.WithAuth(tokenB),
			setup.WithOrg(orgA.Data.ID),
		)
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound,
			"Expected 403 or 404 but got %d", resp.StatusCode)
	})

	t.Run("User A can view own org details", func(t *testing.T) {
		resp := client.GET("/api/v1/organizations/"+orgA.Data.ID,
			setup.WithAuth(tokenA),
			setup.WithOrg(orgA.Data.ID),
		)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestCrossTenantIsolation_PermissionManagement tests that a user from Org A
// cannot assign permissions or roles outside their organization's domain.
func TestCrossTenantIsolation_PermissionManagement(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("TestPass123!"), bcrypt.DefaultCost)

	userA := f.Create(func(u *userEntity.User) {
		u.Username = "perm_user_a"
		u.Email = "perm_a@test.com"
		u.Password = string(hash)
	})

	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": userA.Username,
		"password": "TestPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var la struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&la)
	tokenA := la.Data.AccessToken

	// Create Org A
	resp = client.POST("/api/v1/organizations", map[string]any{
		"name": "Perm Org A",
		"slug": "perm-org-a",
	}, setup.WithAuth(tokenA))
	require.Equal(t, 201, resp.StatusCode)
	var orgA struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&orgA)

	// Create Org B (simulating another org)
	orgBID := uuid.New().String()

	// Grant Casbin policies for User A to manage permissions
	roleA := uuid.New().String()
	server.DB.Create(&roleEntity.Role{
		ID: roleA, Name: "PermAdmin", OrganizationID: &orgA.Data.ID,
	})
	_, _ = server.Enforcer.AddGroupingPolicy(userA.ID, roleA, orgA.Data.ID)
	// Give permission to grant permissions (POST /api/v1/permissions/grant)
	_, _ = server.Enforcer.AddPolicy(roleA, orgA.Data.ID, "/api/v1/permissions/grant", "POST")
	_ = server.Enforcer.SavePolicy()

	t.Run("User A tries to grant permission to global domain", func(t *testing.T) {
		resp := client.POST("/api/v1/permissions/grant", map[string]any{
			"role":   "PermAdmin",
			"path":   "/api/v1/hacked",
			"method": "POST",
			"domain": "global", // Malicious attempt to override domain
		},
			setup.WithAuth(tokenA),
			setup.WithOrg(orgA.Data.ID),
		)

		// The request itself will succeed because User A has access to POST /permissions/grant
		// BUT the domain should be forced to orgA.Data.ID internally.
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify in enforcer that "global" was NOT affected
		hasGlobal, _ := server.Enforcer.HasPolicy("PermAdmin", "global", "/api/v1/hacked", "POST")
		assert.False(t, hasGlobal, "Malicious user successfully injected a global policy!")

		// Verify in enforcer that "Org A" WAS affected (override worked)
		hasOrgA, _ := server.Enforcer.HasPolicy("PermAdmin", orgA.Data.ID, "/api/v1/hacked", "POST")
		assert.True(t, hasOrgA, "Domain was not correctly scoped to the tenant's organization ID")
	})

	t.Run("User A tries to grant permission to Org B domain", func(t *testing.T) {
		resp := client.POST("/api/v1/permissions/grant", map[string]any{
			"role":   "PermAdmin",
			"path":   "/api/v1/hacked-again",
			"method": "GET",
			"domain": orgBID, // Malicious attempt to override domain
		},
			setup.WithAuth(tokenA),
			setup.WithOrg(orgA.Data.ID),
		)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Verify in enforcer that Org B was NOT affected
		hasOrgB, _ := server.Enforcer.HasPolicy("PermAdmin", orgBID, "/api/v1/hacked-again", "GET")
		assert.False(t, hasOrgB, "Malicious user successfully injected a policy into another organization!")

		// Verify in enforcer that "Org A" WAS affected
		hasOrgA, _ := server.Enforcer.HasPolicy("PermAdmin", orgA.Data.ID, "/api/v1/hacked-again", "GET")
		assert.True(t, hasOrgA, "Domain was not correctly scoped to the tenant's organization ID")
	})
}
