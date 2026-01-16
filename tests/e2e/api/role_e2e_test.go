//go:build e2e
// +build e2e

package api

import (
	"testing"

	roleEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// Helper: Create admin user and get token for role tests
func createRoleAdminAndLogin(t *testing.T, server *setup.TestServer) string {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("RoleAdmin123!"), bcrypt.DefaultCost)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "role_admin"
		u.Email = "role_admin@test.com"
		u.Password = string(hash)
	})

	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin")
	server.Enforcer.AddPolicy("role:superadmin", "*", "*")
	server.Enforcer.SavePolicy()

	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": admin.Username,
		"password": "RoleAdmin123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	return loginRes.Data.AccessToken
}

func TestRoleE2E_CreateRole(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createRoleAdminAndLogin(t, server)

	t.Run("Success - Create Role", func(t *testing.T) {
		roleName := "TestRole_" + uuid.New().String()[:8]
		resp := server.Client.POST("/api/v1/roles", map[string]any{
			"name":        roleName,
			"description": "A test role",
		}, setup.WithAuth(adminToken))

		assert.Equal(t, 201, resp.StatusCode)

		var result struct {
			Data struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		assert.Equal(t, roleName, result.Data.Name)
	})

	t.Run("Negative - Duplicate Role Name", func(t *testing.T) {
		roleName := "DuplicateRole"
		server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: roleName})

		resp := server.Client.POST("/api/v1/roles", map[string]any{
			"name": roleName,
		}, setup.WithAuth(adminToken))

		assert.Equal(t, 409, resp.StatusCode)
	})

	t.Run("Negative - Empty Name", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/roles", map[string]any{
			"name": "",
		}, setup.WithAuth(adminToken))

		assert.Equal(t, 422, resp.StatusCode)
	})
}

func TestRoleE2E_DeleteRole(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createRoleAdminAndLogin(t, server)

	t.Run("Success - Delete Role", func(t *testing.T) {
		// Create role to delete
		roleToDelete := &roleEntity.Role{ID: uuid.New().String(), Name: "RoleToDelete"}
		server.DB.Create(roleToDelete)

		resp := server.Client.DELETE("/api/v1/roles/"+roleToDelete.ID, setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Negative - Delete Non-existent", func(t *testing.T) {
		resp := server.Client.DELETE("/api/v1/roles/nonexistent-role-id", setup.WithAuth(adminToken))
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestRoleE2E_GetAllRoles(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createRoleAdminAndLogin(t, server)

	// Create some roles
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Role_List_1"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Role_List_2"})

	t.Run("Success - Get All Roles", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/roles", setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.Data), 2)
	})
}

func TestRoleE2E_DynamicSearch(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createRoleAdminAndLogin(t, server)

	// Create roles for searching
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "SearchableRole_Alpha"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "SearchableRole_Beta"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "OtherRole"})

	t.Run("Success - Search by Name", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/roles/search", map[string]any{
			"filter": map[string]any{
				"name": map[string]any{"type": "contains", "from": "Searchable"},
			},
		}, setup.WithAuth(adminToken))

		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data []struct {
				Name string `json:"name"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.Data), 2)
	})
}
