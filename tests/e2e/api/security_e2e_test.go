//go:build e2e
// +build e2e

package api

import (
	"net/http"
	"testing"

	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestSecurityE2E_AdminBanUser(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)

	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	adminUser := f.Create(func(u *userEntity.User) {
		u.Username = "admin_banner"
		u.Email = "admin@ban.com"
		u.Password = passHash
	})
	server.Enforcer.AddGroupingPolicy(adminUser.ID, "role:superadmin")

	targetUser := f.Create(func(u *userEntity.User) {
		u.Username = "target_user"
		u.Email = "target@ban.com"
		u.Password = passHash
	})
	server.Enforcer.AddGroupingPolicy(targetUser.ID, "role:user")

	loginPayload := map[string]any{"username": targetUser.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&loginRes)
	targetToken := loginRes.Data.AccessToken

	resp = client.GET("/api/v1/users/me", setup.WithAuth(targetToken))
	assert.Equal(t, 200, resp.StatusCode)

	server.DB.Model(&userEntity.User{}).Where("id = ?", targetUser.ID).Update("status", userEntity.UserStatusBanned)

	resp = client.GET("/api/v1/users/me", setup.WithAuth(targetToken))

	assert.Equal(t, 403, resp.StatusCode, "Banned user should be denied access")
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

	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin")
	server.Enforcer.AddPolicy("role:superadmin", "/api/v1/permissions/grant", "POST")
	server.Enforcer.AddPolicy("role:superadmin", "/api/v1/permissions/assign-role", "POST")
	server.Enforcer.AddPolicy("role:superadmin", "/api/v1/roles", "GET")

	server.Enforcer.SavePolicy()

	loginPayload := map[string]any{"username": user.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&loginRes)
	userToken := loginRes.Data.AccessToken

	adminLoginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	respAdmin := client.POST("/api/v1/auth/login", adminLoginPayload)
	var adminLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	respAdmin.JSON(&adminLoginRes)
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

	userRoles, _ := server.Enforcer.GetRolesForUser(user.ID)

	t.Logf("DEBUG: Roles for user %s: %v", user.ID, userRoles)

	hasPermission, _ := server.Enforcer.HasPolicy(roleName, "/api/v1/roles", "GET")

	t.Logf("DEBUG: Role %s has permission GET /api/v1/roles: %v", roleName, hasPermission)

	server.Enforcer.LoadPolicy()

	resp = client.GET("/api/v1/roles", setup.WithAuth(userToken))

	assert.Equal(t, 200, resp.StatusCode)

}

func TestSecurityE2E_TokenRotation(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)

	user := f.Create(func(u *userEntity.User) {
		u.Username = "rotate_user"
		u.Email = "rot@test.com"
		u.Password = string(hash)
	})
	server.Enforcer.AddGroupingPolicy(user.ID, "role:user")

	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": user.Username, "password": "StrongPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	cookies := resp.Cookies()
	var refreshToken1 *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshToken1 = c
			break
		}
	}
	require.NotNil(t, refreshToken1, "Refresh token cookie not found")

	req, _ := http.NewRequest("POST", server.BaseURL+"/api/v1/auth/refresh", nil)
	req.AddCookie(refreshToken1)

	clientWithCookie := &http.Client{}
	respRotate, err := clientWithCookie.Do(req)
	require.NoError(t, err)
	defer respRotate.Body.Close()

	require.Equal(t, 200, respRotate.StatusCode)

	cookies2 := respRotate.Cookies()
	var refreshToken2 *http.Cookie
	for _, c := range cookies2 {
		if c.Name == "refresh_token" {
			refreshToken2 = c
			break
		}
	}
	require.NotNil(t, refreshToken2)
	assert.NotEqual(t, refreshToken1.Value, refreshToken2.Value)

	reqReuse, _ := http.NewRequest("POST", server.BaseURL+"/api/v1/auth/refresh", nil)
	reqReuse.AddCookie(refreshToken1)

	respReuse, err := clientWithCookie.Do(reqReuse)
	require.NoError(t, err)
	defer respReuse.Body.Close()

	assert.Contains(t, []int{401, 400}, respReuse.StatusCode)
}
