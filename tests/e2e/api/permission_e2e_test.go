//go:build e2e
// +build e2e

package api

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPermissionE2E_RoleHierarchy(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "hier_admin"
		u.Email = "hier@admin.com"
		u.Password = passHash
	})

	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin")
	server.Enforcer.AddPolicy("role:superadmin", "*", "*")
	server.Enforcer.SavePolicy()

	loginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&loginRes)
	adminToken := loginRes.Data.AccessToken

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Supervisor"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Intern"})

	grantPayload := map[string]any{
		"role": "Intern", "path": "/api/v1/coffee", "method": "GET",
	}
	resp = client.POST("/api/v1/permissions/grant", grantPayload, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)

	supervisorUser := f.Create(func(u *userEntity.User) { u.Username = "supervisor"; u.Password = passHash })
	internUser := f.Create(func(u *userEntity.User) { u.Username = "intern"; u.Password = passHash })

	client.POST("/api/v1/permissions/assign-role", map[string]any{"user_id": supervisorUser.ID, "role": "Supervisor"}, setup.WithAuth(adminToken))
	client.POST("/api/v1/permissions/assign-role", map[string]any{"user_id": internUser.ID, "role": "Intern"}, setup.WithAuth(adminToken))

	resp = client.POST("/api/v1/auth/login", map[string]any{"username": "supervisor", "password": "StrongPass123!"})
	var supLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&supLoginRes)
	supToken := supLoginRes.Data.AccessToken

	batchReq := model.BatchPermissionCheckRequest{
		Items: []model.PermissionCheckItem{
			{Resource: "/api/v1/coffee", Action: "GET"},
		},
	}
	resp = client.POST("/api/v1/permissions/check-batch", batchReq, setup.WithAuth(supToken))
	require.Equal(t, 200, resp.StatusCode)
	var checkRes struct {
		Data model.BatchPermissionCheckResponse `json:"data"`
	}
	resp.JSON(&checkRes)

	assert.False(t, checkRes.Data.Results["/api/v1/coffee:GET"], "Supervisor should NOT have access yet")

	inheritPayload := map[string]any{
		"child_role":  "Supervisor",
		"parent_role": "Intern",
	}
	resp = client.POST("/api/v1/permissions/inheritance", inheritPayload, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	resp = client.POST("/api/v1/permissions/check-batch", batchReq, setup.WithAuth(supToken))
	require.Equal(t, 200, resp.StatusCode)
	resp.JSON(&checkRes)

	assert.True(t, checkRes.Data.Results["/api/v1/coffee:GET"], "Supervisor SHOULD have access after inheritance")
}

func TestPermissionE2E_BatchCheck(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) { u.Username = "batch_admin"; u.Password = passHash })
	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin")
	server.Enforcer.AddPolicy("role:superadmin", "*", "*")
	server.Enforcer.SavePolicy()

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	var adminRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&adminRes)
	adminToken := adminRes.Data.AccessToken

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Editor"})

	client.POST("/api/v1/permissions/grant", map[string]any{"role": "Editor", "path": "/news", "method": "READ"}, setup.WithAuth(adminToken))
	client.POST("/api/v1/permissions/grant", map[string]any{"role": "Editor", "path": "/news", "method": "WRITE"}, setup.WithAuth(adminToken))

	user := f.Create(func(u *userEntity.User) { u.Username = "editor_user"; u.Password = passHash })
	client.POST("/api/v1/permissions/assign-role", map[string]any{"user_id": user.ID, "role": "Editor"}, setup.WithAuth(adminToken))

	resp = client.POST("/api/v1/auth/login", map[string]any{"username": user.Username, "password": "StrongPass123!"})
	var userRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&userRes)
	userToken := userRes.Data.AccessToken

	roles, _ := server.Enforcer.GetRolesForUser(user.ID)
	t.Logf("DEBUG: Roles for user %s: %v", user.ID, roles)

	req := model.BatchPermissionCheckRequest{
		Items: []model.PermissionCheckItem{
			{Resource: "/news", Action: "READ"},
			{Resource: "/news", Action: "WRITE"},
			{Resource: "/news", Action: "DELETE"},
			{Resource: "/admin", Action: "GET"},
		},
	}

	resp = client.POST("/api/v1/permissions/check-batch", req, setup.WithAuth(userToken))
	require.Equal(t, 200, resp.StatusCode)

	var checkRes struct {
		Data model.BatchPermissionCheckResponse `json:"data"`
	}
	err := resp.JSON(&checkRes)
	require.NoError(t, err)

	assert.True(t, checkRes.Data.Results["/news:READ"])
	assert.True(t, checkRes.Data.Results["/news:WRITE"])
	assert.False(t, checkRes.Data.Results["/news:DELETE"])
	assert.False(t, checkRes.Data.Results["/admin:GET"])
}

func TestPermissionE2E_RevokeRole(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "revoke_admin"
		u.Email = "revoke@admin.com"
		u.Password = passHash
	})

	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin")
	server.Enforcer.AddPolicy("role:superadmin", "*", "*")
	server.Enforcer.SavePolicy()

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	adminToken := loginRes.Data.AccessToken

	// Create role and user
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "RevokeTestRole"})
	user := f.Create(func(u *userEntity.User) { u.Username = "revoke_user"; u.Password = passHash })

	// Assign role
	resp = client.POST("/api/v1/permissions/assign-role", map[string]any{
		"user_id": user.ID,
		"role":    "RevokeTestRole",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	// Verify role assigned
	roles, _ := server.Enforcer.GetRolesForUser(user.ID)
	assert.Contains(t, roles, "RevokeTestRole")

	t.Run("Success - Revoke Role", func(t *testing.T) {
		resp := client.DELETE("/api/v1/permissions/revoke-role", map[string]any{
			"user_id": user.ID,
			"role":    "RevokeTestRole",
		}, setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		// Verify role revoked
		rolesAfter, _ := server.Enforcer.GetRolesForUser(user.ID)
		assert.NotContains(t, rolesAfter, "RevokeTestRole")
	})
}

func TestPermissionE2E_RemoveInheritance(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "remove_inherit_admin"
		u.Email = "remove_inherit@admin.com"
		u.Password = passHash
	})

	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin")
	server.Enforcer.AddPolicy("role:superadmin", "*", "*")
	server.Enforcer.SavePolicy()

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	adminToken := loginRes.Data.AccessToken

	// Create roles
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "ParentRole"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "ChildRole"})

	// Grant permission to child
	client.POST("/api/v1/permissions/grant", map[string]any{
		"role": "ChildRole", "path": "/api/v1/inherited", "method": "GET",
	}, setup.WithAuth(adminToken))

	// Add inheritance - Parent inherits from Child
	resp = client.POST("/api/v1/permissions/inheritance", map[string]any{
		"child_role":  "ParentRole",
		"parent_role": "ChildRole",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	// Verify inheritance
	ok, _ := server.Enforcer.Enforce("ParentRole", "/api/v1/inherited", "GET")
	assert.True(t, ok, "Parent should have access via inheritance")

	t.Run("Success - Remove Inheritance", func(t *testing.T) {
		resp := client.DELETE("/api/v1/permissions/inheritance", map[string]any{
			"child_role":  "ParentRole",
			"parent_role": "ChildRole",
		}, setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		// Verify inheritance removed
		ok, _ := server.Enforcer.Enforce("ParentRole", "/api/v1/inherited", "GET")
		assert.False(t, ok, "Parent should NOT have access after inheritance removed")
	})
}
