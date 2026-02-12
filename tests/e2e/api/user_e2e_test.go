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
