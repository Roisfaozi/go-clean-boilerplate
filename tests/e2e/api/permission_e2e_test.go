package api

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
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

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)

	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	loginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
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
	err = resp.JSON(&supLoginRes)
	require.NoError(t, err)
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
	err = resp.JSON(&checkRes)
	require.NoError(t, err)

	assert.False(t, checkRes.Data.Results["/api/v1/coffee:GET"], "Supervisor should NOT have access yet")

	inheritPayload := map[string]any{
		"child_role":  "Supervisor",
		"parent_role": "Intern",
	}
	resp = client.POST("/api/v1/permissions/inheritance", inheritPayload, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	resp = client.POST("/api/v1/permissions/check-batch", batchReq, setup.WithAuth(supToken))
	require.Equal(t, 200, resp.StatusCode)
	err = resp.JSON(&checkRes)
	require.NoError(t, err)

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
	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	var adminRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&adminRes)
	require.NoError(t, err)
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
	err = resp.JSON(&userRes)
	require.NoError(t, err)
	userToken := userRes.Data.AccessToken

	roles, _ := server.Enforcer.GetRolesForUser(user.ID, "global")
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
	err = resp.JSON(&checkRes)
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

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	adminToken := loginRes.Data.AccessToken

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "RevokeTestRole"})
	user := f.Create(func(u *userEntity.User) { u.Username = "revoke_user"; u.Password = passHash })

	resp = client.POST("/api/v1/permissions/assign-role", map[string]any{
		"user_id": user.ID,
		"role":    "RevokeTestRole",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	roles, _ := server.Enforcer.GetRolesForUser(user.ID, "global")
	assert.Contains(t, roles, "RevokeTestRole")

	t.Run("Success - Revoke Role", func(t *testing.T) {
		resp := client.DELETE("/api/v1/permissions/revoke-role", map[string]any{
			"user_id": user.ID,
			"role":    "RevokeTestRole",
		}, setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		rolesAfter, _ := server.Enforcer.GetRolesForUser(user.ID, "global")
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

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	adminToken := loginRes.Data.AccessToken

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "ParentRole"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "ChildRole"})

	client.POST("/api/v1/permissions/grant", map[string]any{
		"role": "ChildRole", "path": "/api/v1/inherited", "method": "GET",
	}, setup.WithAuth(adminToken))

	resp = client.POST("/api/v1/permissions/inheritance", map[string]any{
		"child_role":  "ParentRole",
		"parent_role": "ChildRole",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	ok, _ := server.Enforcer.Enforce("ParentRole", "global", "/api/v1/inherited", "GET")
	assert.True(t, ok, "Parent should have access via inheritance")

	t.Run("Success - Remove Inheritance", func(t *testing.T) {
		resp := client.DELETE("/api/v1/permissions/inheritance", map[string]any{
			"child_role":  "ParentRole",
			"parent_role": "ChildRole",
		}, setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		ok, _ := server.Enforcer.Enforce("ParentRole", "global", "/api/v1/inherited", "GET")
		assert.False(t, ok, "Parent should NOT have access after inheritance removed")
	})
}

func TestPermissionE2E_Security_PostRoleDeletion_PermissionsRevoked(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "post_del_admin"
		u.Email = "postdel@admin.com"
		u.Password = passHash
	})
	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	require.Equal(t, 200, resp.StatusCode)
	var adminLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&adminLoginRes)
	require.NoError(t, err)
	adminToken := adminLoginRes.Data.AccessToken

	roleID := uuid.New().String()
	server.DB.Create(&roleEntity.Role{ID: roleID, Name: "EphemeralRole"})
	_, err = server.Enforcer.AddPolicy("EphemeralRole", "global", "/api/v1/secret", "GET")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	targetUser := f.Create(func(u *userEntity.User) { u.Username = "ephemeral_user"; u.Password = passHash })
	client.POST("/api/v1/permissions/assign-role", map[string]any{
		"user_id": targetUser.ID,
		"role":    "EphemeralRole",
	}, setup.WithAuth(adminToken))

	resp = client.POST("/api/v1/auth/login", map[string]any{"username": targetUser.Username, "password": "StrongPass123!"})
	var targetLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	err = resp.JSON(&targetLoginRes)
	require.NoError(t, err)
	targetToken := targetLoginRes.Data.AccessToken

	roles, _ := server.Enforcer.GetRolesForUser(targetUser.ID, "global")
	assert.Contains(t, roles, "EphemeralRole", "User should have role before deletion")

	resp = client.DELETE("/api/v1/roles/"+roleID, setup.WithAuth(adminToken))
	assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 204, "Role deletion should succeed")

	policies, _ := server.Enforcer.GetFilteredPolicy(0, "EphemeralRole")
	assert.Empty(t, policies, "All Casbin policies for deleted role must be removed")

	rolesAfter, _ := server.Enforcer.GetRolesForUser(targetUser.ID, "global")
	assert.NotContains(t, rolesAfter, "EphemeralRole", "User's role assignment must be removed after role deletion")

	_ = targetToken
	hasAccess, _ := server.Enforcer.Enforce(targetUser.ID, "global", "/api/v1/secret", "GET")
	assert.False(t, hasAccess, "User must NOT have access after role deletion")
}

func TestPermissionE2E_AccessRightAssignment(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "ar_admin"
		u.Email = "ar@admin.com"
		u.Password = passHash
	})

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	loginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	adminToken := loginRes.Data.AccessToken

	ep1 := entity.Endpoint{Path: "/api/test", Method: "GET"}
	ep2 := entity.Endpoint{Path: "/api/test", Method: "POST"}
	server.DB.Create(&ep1)
	server.DB.Create(&ep2)

	ar := entity.AccessRight{
		Name:      "Test Access Right",
		Endpoints: []entity.Endpoint{ep1, ep2},
	}
	server.DB.Create(&ar)

	roleName := "TestRole"

	t.Run("GetRoleAccessRights - Unassigned", func(t *testing.T) {
		resp := client.GET("/api/v1/permissions/roles/"+roleName+"/access-rights", setup.WithAuth(adminToken))
		require.Equal(t, 200, resp.StatusCode)

		var res struct {
			Data []model.RoleAccessRightStatus `json:"data"`
		}
		err = resp.JSON(&res)
		require.NoError(t, err)

		found := false
		for _, s := range res.Data {
			if s.ID == ar.ID {
				found = true
				assert.False(t, s.Assigned)
				assert.False(t, s.Partial)
			}
		}
		assert.True(t, found, "Access right should be in the list")
	})

	t.Run("AssignAccessRight", func(t *testing.T) {
		payload := map[string]any{
			"role":            roleName,
			"access_right_id": ar.ID,
		}
		resp := client.POST("/api/v1/permissions/assign-access-right", payload, setup.WithAuth(adminToken))
		require.Equal(t, 200, resp.StatusCode)

		ok, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "GET")
		assert.True(t, ok)
		ok2, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "POST")
		assert.True(t, ok2)
	})

	t.Run("GetRoleAccessRights - Assigned", func(t *testing.T) {
		resp := client.GET("/api/v1/permissions/roles/"+roleName+"/access-rights", setup.WithAuth(adminToken))
		require.Equal(t, 200, resp.StatusCode)

		var res struct {
			Data []model.RoleAccessRightStatus `json:"data"`
		}
		err = resp.JSON(&res)
		require.NoError(t, err)

		for _, s := range res.Data {
			if s.ID == ar.ID {
				assert.True(t, s.Assigned)
			}
		}
	})

	t.Run("RevokeAccessRight", func(t *testing.T) {
		payload := map[string]any{
			"role":            roleName,
			"access_right_id": ar.ID,
		}
		resp := client.DELETE("/api/v1/permissions/revoke-access-right", payload, setup.WithAuth(adminToken))
		require.Equal(t, 200, resp.StatusCode)

		ok, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "GET")
		assert.False(t, ok)
		ok2, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "POST")
		assert.False(t, ok2)
	})
}

func TestAccessRightsFlowE2E(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("AdminPass123!"), bcrypt.DefaultCost)
	admin := f.Create(func(u *userEntity.User) {
		u.Username = "flow_admin"
		u.Password = string(hash)
	})
	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	require.NoError(t, err)
	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": admin.Username,
		"password": "AdminPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	adminToken := loginRes.Data.AccessToken

	resp = client.POST("/api/v1/access-rights", map[string]any{
		"name":        "User Management",
		"description": "Operations related to user accounts",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var arRes struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	err = resp.JSON(&arRes)
	require.NoError(t, err)
	arID := arRes.Data.ID

	resp = client.POST("/api/v1/endpoints", map[string]any{
		"path":   "/api/v1/users",
		"method": "GET",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var epRes struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	err = resp.JSON(&epRes)
	require.NoError(t, err)
	epID := epRes.Data.ID

	resp = client.POST("/api/v1/access-rights/link", map[string]any{
		"access_right_id": arID,
		"endpoint_id":     epID,
	}, setup.WithAuth(adminToken))
	assert.Equal(t, 200, resp.StatusCode)

	roleID := uuid.New().String()
	server.DB.Create(&roleEntity.Role{ID: roleID, Name: "UserManager"})

	resp = client.POST("/api/v1/permissions/grant", map[string]any{
		"role":   "UserManager",
		"path":   "/api/v1/users",
		"method": "GET",
	}, setup.WithAuth(adminToken))
	assert.Equal(t, 201, resp.StatusCode)

	user := f.Create(func(u *userEntity.User) { u.Username = "manager_user"; u.Password = string(hash) })
	client.POST("/api/v1/permissions/assign-role", map[string]any{
		"user_id": user.ID,
		"role":    "UserManager",
	}, setup.WithAuth(adminToken))

	ok, err := server.Enforcer.Enforce(user.ID, "global", "/api/v1/users", "GET")
	require.NoError(t, err)
	assert.True(t, ok, "User should have access to linked endpoint")
}

func TestSecurityE2E_DynamicRBAC(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	user := f.Create(func(u *userEntity.User) {
		u.Username = "rbac_user"
		u.Email = "rbac@test.com"
		u.Password = passHash
	})

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "super_admin"
		u.Email = "super@admin.com"
		u.Password = passHash
	})

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "/api/v1/permissions/grant", "POST")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "/api/v1/permissions/assign-role", "POST")
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", "global", "/api/v1/roles", "GET")
	require.NoError(t, err)

	err = server.Enforcer.SavePolicy()
	require.NoError(t, err)

	loginPayload := map[string]any{"username": user.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = resp.JSON(&loginRes)
	require.NoError(t, err)
	userToken := loginRes.Data.AccessToken

	adminLoginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	respAdmin := client.POST("/api/v1/auth/login", adminLoginPayload)
	var adminLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	err = respAdmin.JSON(&adminLoginRes)
	require.NoError(t, err)
	adminToken := adminLoginRes.Data.AccessToken

	resp = client.GET("/api/v1/roles", setup.WithAuth(userToken))
	assert.Equal(t, 403, resp.StatusCode)

	roleName := "DynamicViewer"

	server.DB.Create(&roleEntity.Role{Name: roleName})

	grantPayload := map[string]any{
		"role":   roleName,
		"path":   "/api/v1/roles",
		"method": "GET",
	}
	resp = client.POST("/api/v1/permissions/grant", grantPayload, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)

	assignPayload := map[string]any{

		"user_id": user.ID,

		"role": roleName,
	}

	resp = client.POST("/api/v1/permissions/assign-role", assignPayload, setup.WithAuth(adminToken))

	require.Equal(t, 200, resp.StatusCode)

	userRoles, _ := server.Enforcer.GetRolesForUser(user.ID, "global")

	t.Logf("DEBUG: Roles for user %s: %v", user.ID, userRoles)

	hasPermission, _ := server.Enforcer.HasPolicy(roleName, "global", "/api/v1/roles", "GET")

	t.Logf("DEBUG: Role %s has permission GET /api/v1/roles: %v", roleName, hasPermission)

	err = server.Enforcer.LoadPolicy()
	require.NoError(t, err)

	resp = client.GET("/api/v1/roles", setup.WithAuth(userToken))

	assert.Equal(t, 200, resp.StatusCode)

}
