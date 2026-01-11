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

// Scenario 1: Admin Bans User (Real-time enforcement)
func TestSecurityE2E_AdminBanUser(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client
	
	// Create Users
	f := fixtures.NewUserFactory(server.DB)
	
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	// Admin user (not strictly needed for the Ban action since we simulate DB update, 
	// but good for realism if we expanded)
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

	// 2. Login as Target User
	loginPayload := map[string]any{"username": targetUser.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)
	
	var loginRes struct { Data struct { AccessToken string `json:"access_token"` } }
	resp.JSON(&loginRes)
	targetToken := loginRes.Data.AccessToken

	// 3. Target User Accesses Protected Route (Should Succeed)
	resp = client.GET("/api/v1/users/me", setup.WithAuth(targetToken))
	assert.Equal(t, 200, resp.StatusCode)

	// 4. Admin Bans Target User via DB simulation
	// In a real scenario, Admin would call PUT /api/v1/users/:id/status
	// But direct DB update is sufficient to test Middleware enforcement.
	server.DB.Model(&userEntity.User{}).Where("id = ?", targetUser.ID).Update("status", userEntity.UserStatusBanned)
	
	// 5. Target User Tries to Access Protected Route (Should Fail)
	resp = client.GET("/api/v1/users/me", setup.WithAuth(targetToken))
	// Expect 403 Forbidden (Middleware returns Forbidden for banned users)
	assert.Equal(t, 403, resp.StatusCode, "Banned user should be denied access")
}

// Scenario 2: Dynamic RBAC Lifecycle
func TestSecurityE2E_DynamicRBAC(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)
	
	// 1. Create User
	user := f.Create(func(u *userEntity.User) {
		u.Username = "rbac_user"
		u.Email = "rbac@test.com"
		u.Password = passHash
	})
	
	// 2. Create Admin (to grant permissions)
	admin := f.Create(func(u *userEntity.User) {
		u.Username = "super_admin"
		u.Email = "super@admin.com"
		u.Password = passHash
	})
	
	// Ensure policies are persistent
	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin")
	server.Enforcer.AddPolicy("role:superadmin", "/api/v1/permissions/grant", "POST")
	server.Enforcer.AddPolicy("role:superadmin", "/api/v1/permissions/assign-role", "POST")
	server.Enforcer.AddPolicy("role:superadmin", "/api/v1/roles", "GET") // Needed for verification later? No, verification is by user.
	// Save just in case
	server.Enforcer.SavePolicy()

	// Login User
	loginPayload := map[string]any{"username": user.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	var loginRes struct { Data struct { AccessToken string `json:"access_token"` } }
	resp.JSON(&loginRes)
	userToken := loginRes.Data.AccessToken

	// Login Admin
	adminLoginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	respAdmin := client.POST("/api/v1/auth/login", adminLoginPayload)
	var adminLoginRes struct { Data struct { AccessToken string `json:"access_token"` } }
	respAdmin.JSON(&adminLoginRes)
	adminToken := adminLoginRes.Data.AccessToken

	// 3. Try to Access Admin Route (Should Fail)
	// /api/v1/roles is usually restricted to admin/superadmin
	resp = client.GET("/api/v1/roles", setup.WithAuth(userToken))
	assert.Equal(t, 403, resp.StatusCode)

	// 4. Admin Grants Permission dynamically via API
	// We create a new role and give it access
	roleName := "DynamicViewer"
	
	// Create Role via DB (Metadata)
	server.DB.Create(&roleEntity.Role{Name: roleName})
	
	// Grant "GET /api/v1/roles" to "DynamicViewer" using API
	// This ensures Enforcer in memory is updated
	grantPayload := map[string]any{
		"role": roleName,
		"path": "/api/v1/roles",
		"method": "GET",
	}
	resp = client.POST("/api/v1/permissions/grant", grantPayload, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	
		// Assign "DynamicViewer" to User via API
	
		assignPayload := map[string]any{
	
			"user_id": user.ID,
	
			"role": roleName,
	
		}
	
		resp = client.POST("/api/v1/permissions/assign-role", assignPayload, setup.WithAuth(adminToken))
	
		require.Equal(t, 200, resp.StatusCode)
	
	
	
		// DEBUG: Inspect Enforcer State
	
		userRoles, _ := server.Enforcer.GetRolesForUser(user.ID)
	
		t.Logf("DEBUG: Roles for user %s: %v", user.ID, userRoles)
	
		
	
		hasPermission, _ := server.Enforcer.HasPolicy(roleName, "/api/v1/roles", "GET")
	
		t.Logf("DEBUG: Role %s has permission GET /api/v1/roles: %v", roleName, hasPermission)
	
	
	
		// Explicitly reload policy to rule out cache issues
	
		server.Enforcer.LoadPolicy()
	
	
	
		// 5. Retry Access (Should Succeed immediately)
	
		resp = client.GET("/api/v1/roles", setup.WithAuth(userToken))
	
		assert.Equal(t, 200, resp.StatusCode)
	
	}

// Scenario 3: Token Rotation Security
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

	// Extract Cookies
	cookies := resp.Cookies()
	var refreshToken1 *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshToken1 = c
			break
		}
	}
	require.NotNil(t, refreshToken1, "Refresh token cookie not found")

	// 2. Rotate Token (RT1 -> RT2)
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

	// 3. Attempt Reuse RT1 (Should Fail)
	reqReuse, _ := http.NewRequest("POST", server.BaseURL+"/api/v1/auth/refresh", nil)
	reqReuse.AddCookie(refreshToken1) // OLD token
	
	respReuse, err := clientWithCookie.Do(reqReuse)
	require.NoError(t, err)
	defer respReuse.Body.Close()

	// Should fail (401 or 400)
	assert.Contains(t, []int{401, 400}, respReuse.StatusCode)
}
