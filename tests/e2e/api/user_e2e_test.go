//go:build e2e
// +build e2e

package api

import (
	"testing"

	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserE2E_GetAllUsers(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := setup.CreateAdminAndLogin(t, server)

	// Create some users
	f := fixtures.NewUserFactory(server.DB)
	f.Create(func(u *userEntity.User) { u.Username = "user_list_1"; u.Email = "list1@test.com" })
	f.Create(func(u *userEntity.User) { u.Username = "user_list_2"; u.Email = "list2@test.com" })

	t.Run("Success - Get All Users", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/users", setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data []struct {
				ID       string `json:"id"`
				Username string `json:"username"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.Data), 2)
	})

	t.Run("Negative - Unauthorized", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/users")
		assert.Equal(t, 401, resp.StatusCode)
	})
}

func TestUserE2E_GetUserByID(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := setup.CreateAdminAndLogin(t, server)

	// Create target user
	f := fixtures.NewUserFactory(server.DB)
	targetUser := f.Create(func(u *userEntity.User) {
		u.Username = "target_by_id"
		u.Email = "target_byid@test.com"
		u.Name = "Target User"
	})

	t.Run("Success - Get User By ID", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/users/"+targetUser.ID, setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data struct {
				ID       string `json:"id"`
				Username string `json:"username"`
				Name     string `json:"name"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		assert.Equal(t, targetUser.ID, result.Data.ID)
		assert.Equal(t, "Target User", result.Data.Name)
	})

	t.Run("Negative - Not Found", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/users/nonexistent-id-12345", setup.WithAuth(adminToken))
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestUserE2E_DeleteUser(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := setup.CreateAdminAndLogin(t, server)

	// Create user to delete
	f := fixtures.NewUserFactory(server.DB)
	userToDelete := f.Create(func(u *userEntity.User) {
		u.Username = "user_to_delete"
		u.Email = "delete@test.com"
	})

	t.Run("Success - Delete User", func(t *testing.T) {
		resp := server.Client.DELETE("/api/v1/users/"+userToDelete.ID, setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		// Verify user is gone
		resp = server.Client.GET("/api/v1/users/"+userToDelete.ID, setup.WithAuth(adminToken))
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("Negative - Delete Non-existent", func(t *testing.T) {
		resp := server.Client.DELETE("/api/v1/users/nonexistent-id", setup.WithAuth(adminToken))
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestUserE2E_UpdateStatus(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := setup.CreateAdminAndLogin(t, server)

	// Create user and get their token
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("UserPass123!"), bcrypt.DefaultCost)
	targetUser := f.Create(func(u *userEntity.User) {
		u.Username = "status_user"
		u.Email = "status@test.com"
		u.Password = string(hash)
		u.Status = userEntity.UserStatusActive
	})

	// Login target user
	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": targetUser.Username,
		"password": "UserPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var userLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&userLoginRes)
	userToken := userLoginRes.Data.AccessToken

	t.Run("Success - Ban User", func(t *testing.T) {
		// Admin bans user
		resp := server.Client.PATCH("/api/v1/users/"+targetUser.ID+"/status",
			map[string]any{"status": "banned"},
			setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		// Banned user tries to access protected route
		// Note: When banned, user's sessions are revoked, so token becomes invalid (401)
		resp = server.Client.GET("/api/v1/users/me", setup.WithAuth(userToken))
		assert.Equal(t, 401, resp.StatusCode, "Banned user's token should be invalidated")
	})

	t.Run("Success - Reactivate User", func(t *testing.T) {
		// Admin reactivates user
		resp := server.Client.PATCH("/api/v1/users/"+targetUser.ID+"/status",
			map[string]any{"status": "active"},
			setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Negative - Invalid Status", func(t *testing.T) {
		resp := server.Client.PATCH("/api/v1/users/"+targetUser.ID+"/status",
			map[string]any{"status": "invalid_status"},
			setup.WithAuth(adminToken))
		assert.Equal(t, 422, resp.StatusCode)
	})
}

func TestUserE2E_Security_IDOR_UpdateOtherUser(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	hash, _ := bcrypt.GenerateFromPassword([]byte("UserPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	f := fixtures.NewUserFactory(server.DB)

	// Create attacker (regular user, no admin role)
	attacker := f.Create(func(u *userEntity.User) {
		u.Username = "idor_attacker"
		u.Email = "attacker@test.com"
		u.Password = passHash
		u.Status = userEntity.UserStatusActive
	})

	// Create victim user
	victim := f.Create(func(u *userEntity.User) {
		u.Username = "idor_victim"
		u.Email = "victim@test.com"
		u.Password = passHash
		u.Status = userEntity.UserStatusActive
	})

	// Login as attacker
	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": attacker.Username,
		"password": "UserPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	attackerToken := loginRes.Data.AccessToken

	t.Run("Security - IDOR: Attacker cannot delete victim", func(t *testing.T) {
		// Attacker (regular user) should NOT be able to delete another user via admin endpoint
		resp := server.Client.DELETE("/api/v1/users/"+victim.ID,
			setup.WithAuth(attackerToken))
		assert.True(t, resp.StatusCode == 403 || resp.StatusCode == 401,
			"Non-admin user must NOT be able to delete another user's account (got %d)", resp.StatusCode)
	})

	t.Run("Security - IDOR: Attacker cannot change victim status", func(t *testing.T) {
		// Attacker (regular user) should NOT be able to change another user's status
		resp := server.Client.PATCH("/api/v1/users/"+victim.ID+"/status",
			map[string]any{"status": "banned"},
			setup.WithAuth(attackerToken))
		assert.True(t, resp.StatusCode == 403 || resp.StatusCode == 401,
			"Non-admin user must NOT be able to change another user's status (got %d)", resp.StatusCode)
	})

	t.Run("Positive - User can update own profile via /me", func(t *testing.T) {
		updatePayload := map[string]any{
			"name": "Attacker Updated Own",
		}
		resp := server.Client.PUT("/api/v1/users/me",
			updatePayload,
			setup.WithAuth(attackerToken))
		assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 422,
			"Self-update via /me returns 200 (success) or 422 (validation), got %d", resp.StatusCode)
	})
}

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
	server.Enforcer.AddGroupingPolicy(adminUser.ID, "role:superadmin", "global")

	targetUser := f.Create(func(u *userEntity.User) {
		u.Username = "target_user"
		u.Email = "target@ban.com"
		u.Password = passHash
	})
	server.Enforcer.AddGroupingPolicy(targetUser.ID, "role:user", "global")

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

func TestUserE2E_DynamicSearch(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "search_admin"
		u.Email = "search@admin.com"
		u.Password = passHash
	})
	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	server.Enforcer.SavePolicy()

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&loginRes)
	token := loginRes.Data.AccessToken

	f.Create(func(u *userEntity.User) {
		u.Name = "Alice Wonderland"
		u.Email = "alice@test.com"
		u.Username = "alice_w"
	})
	f.Create(func(u *userEntity.User) { u.Name = "Bob Builder"; u.Email = "bob@test.com"; u.Username = "bob_b" })
	f.Create(func(u *userEntity.User) {
		u.Name = "Charlie Chocolate"
		u.Email = "charlie@test.com"
		u.Username = "charlie_c"
	})

	searchPayload := map[string]any{
		"filter": map[string]any{
			"email": map[string]any{"type": "contains", "from": "alice"},
		},
	}
	resp = client.POST("/api/v1/users/search", searchPayload, setup.WithAuth(token))
	require.Equal(t, 200, resp.StatusCode)

	var userSearchRes struct {
		Data []struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"data"`
		Paging struct {
			Total int64 `json:"total"`
		} `json:"paging"`
	}
	resp.JSON(&userSearchRes)

	if userSearchRes.Paging.Total == 0 {
		t.Logf("DEBUG: Search Result Empty. Payload: %+v. Response: %+v", searchPayload, userSearchRes)
	}

	assert.Equal(t, int64(1), userSearchRes.Paging.Total)
	if len(userSearchRes.Data) > 0 {
		assert.Equal(t, "alice@test.com", userSearchRes.Data[0].Email)
	}

	sortPayload := map[string]any{
		"sort": []map[string]any{
			{"colId": "name", "sort": "desc"},
		},
		"page_size": 5,
	}
	resp = client.POST("/api/v1/users/search", sortPayload, setup.WithAuth(token))
	require.Equal(t, 200, resp.StatusCode)

	var sortedRes struct {
		Data []struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	resp.JSON(&sortedRes)

	assert.NotEmpty(t, sortedRes.Data)
}
