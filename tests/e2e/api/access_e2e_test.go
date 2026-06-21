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

// Helper: Create admin user for access tests
func createAccessAdminAndLogin(t *testing.T, server *setup.TestServer) string {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("AccessAdmin123!"), bcrypt.DefaultCost)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "access_admin"
		u.Email = "access_admin@test.com"
		u.Password = string(hash)
	})

	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	server.Enforcer.SavePolicy()

	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": admin.Username,
		"password": "AccessAdmin123!",
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

func TestAccessE2E_AccessRightsCRUD(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createAccessAdminAndLogin(t, server)
	var createdID string

	t.Run("Success - Create Access Right", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/access-rights", map[string]any{
			"name":        "manage_reports",
			"description": "Can manage reports",
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
		createdID = result.Data.ID
		assert.Equal(t, "manage_reports", result.Data.Name)
	})

	t.Run("Success - Get All Access Rights", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/access-rights", setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data struct {
				Data []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"data"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.Data.Data), 1)
	})

	t.Run("Success - Delete Access Right", func(t *testing.T) {
		resp := server.Client.DELETE("/api/v1/access-rights/"+createdID, setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Negative - Delete Non-existent", func(t *testing.T) {
		resp := server.Client.DELETE("/api/v1/access-rights/nonexistent-id", setup.WithAuth(adminToken))
		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestAccessE2E_EndpointsCRUD(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createAccessAdminAndLogin(t, server)
	var createdID string

	t.Run("Success - Create Endpoint", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/endpoints", map[string]any{
			"path":   "/api/v1/reports",
			"method": "GET",
		}, setup.WithAuth(adminToken))

		assert.Equal(t, 201, resp.StatusCode)

		var result struct {
			Data struct {
				ID     string `json:"id"`
				Path   string `json:"path"`
				Method string `json:"method"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		createdID = result.Data.ID
		assert.Equal(t, "/api/v1/reports", result.Data.Path)
	})

	t.Run("Success - Delete Endpoint", func(t *testing.T) {
		resp := server.Client.DELETE("/api/v1/endpoints/"+createdID, setup.WithAuth(adminToken))
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestAccessE2E_LinkEndpointToAccessRight(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createAccessAdminAndLogin(t, server)

	// Create Access Right
	resp := server.Client.POST("/api/v1/access-rights", map[string]any{
		"name": "link_test_access",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var arResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&arResult)
	accessRightID := arResult.Data.ID

	// Create Endpoint
	resp = server.Client.POST("/api/v1/endpoints", map[string]any{
		"path":   "/api/v1/linked-resource",
		"method": "POST",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var epResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&epResult)
	endpointID := epResult.Data.ID

	t.Run("Success - Link Endpoint to Access Right", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/access-rights/link", map[string]any{
			"access_right_id": accessRightID,
			"endpoint_id":     endpointID,
		}, setup.WithAuth(adminToken))

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Negative - Invalid IDs", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/access-rights/link", map[string]any{
			"access_right_id": "",
			"endpoint_id":     "",
		}, setup.WithAuth(adminToken))

		assert.Equal(t, 422, resp.StatusCode)
	})
}

func TestAccessE2E_DynamicSearch(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createAccessAdminAndLogin(t, server)

	// Create some data
	server.Client.POST("/api/v1/access-rights", map[string]any{
		"name": "searchable_access_alpha",
	}, setup.WithAuth(adminToken))
	server.Client.POST("/api/v1/access-rights", map[string]any{
		"name": "searchable_access_beta",
	}, setup.WithAuth(adminToken))

	server.Client.POST("/api/v1/endpoints", map[string]any{
		"path": "/api/v1/searchable", "method": "GET",
	}, setup.WithAuth(adminToken))
	server.Client.POST("/api/v1/endpoints", map[string]any{
		"path": "/api/v1/searchable", "method": "POST",
	}, setup.WithAuth(adminToken))

	t.Run("Success - Search Access Rights", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/access-rights/search", map[string]any{
			"filter": map[string]any{
				"name": map[string]any{"type": "contains", "from": "searchable"},
			},
		}, setup.WithAuth(adminToken))

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Success - Search Endpoints", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/endpoints/search", map[string]any{
			"filter": map[string]any{
				"path": map[string]any{"type": "contains", "from": "searchable"},
			},
		}, setup.WithAuth(adminToken))

		assert.Equal(t, 200, resp.StatusCode)
	})
}
