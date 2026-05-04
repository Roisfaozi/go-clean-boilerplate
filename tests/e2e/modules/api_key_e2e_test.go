//go:build e2e
// +build e2e

package modules

import (
	"net/http"
	"testing"

	apiKeyModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	integrationSetup "github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyE2E_LifecycleAndAccess(t *testing.T) {
	// 1. Setup Environment
	server := setup.SetupTestServer(t)
	defer server.Server.Close()

	// 2. Create Global Org and Superadmin
	server.DB.Exec("INSERT INTO organizations (id, name, slug, owner_id, status) VALUES (?, ?, ?, ?, ?)", "global", "Global", "global", "system", "active")

	admin := integrationSetup.CreateTestUser(t, server.DB, "api_admin", "admin@api.com", "Password123!", "global")
	// Make admin owner of global
	server.DB.Exec("UPDATE organization_members SET role_id = ? WHERE organization_id = ? AND user_id = ?", "owner", "global", admin.ID)

	_, err := server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	require.NoError(t, err)
	_ = server.Enforcer.LoadPolicy()

	// 3. Login
	loginResp := server.Client.POST("/api/v1/auth/login", map[string]string{
		"username": "api_admin",
		"password": "Password123!",
	})
	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginData struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	_ = loginResp.JSON(&loginData)
	server.Client.Token = loginData.Data.AccessToken

	// 4. Create API Keys with different scopes
	createReq := apiKeyModel.CreateApiKeyRequest{
		Name:   "E2E Read Key",
		Scopes: []string{"project:view"},
	}
	createResp := server.Client.POST("/api/v1/api-keys", createReq, func(r *http.Request) {
		r.Header.Set("X-Organization-ID", "global")
	})
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createData struct {
		Data apiKeyModel.CreateApiKeyResponse `json:"data"`
	}
	_ = createResp.JSON(&createData)
	readOnlyAPIKey := createData.Data.Key
	readOnlyAPIKeyID := createData.Data.ID
	require.NotEmpty(t, readOnlyAPIKey)

	createManageResp := server.Client.POST("/api/v1/api-keys", apiKeyModel.CreateApiKeyRequest{
		Name:   "E2E Manage Key",
		Scopes: []string{"project:manage"},
	}, func(r *http.Request) {
		r.Header.Set("X-Organization-ID", "global")
	})
	require.Equal(t, http.StatusCreated, createManageResp.StatusCode)

	var manageData struct {
		Data apiKeyModel.CreateApiKeyResponse `json:"data"`
	}
	_ = createManageResp.JSON(&manageData)
	manageAPIKey := manageData.Data.Key
	manageAPIKeyID := manageData.Data.ID
	require.NotEmpty(t, manageAPIKey)

	// 5. Use API Keys
	server.Client.Token = ""

	t.Run("API key is rejected on session-only endpoint", func(t *testing.T) {
		meResp := server.Client.GET("/api/v1/users/me", func(r *http.Request) {
			r.Header.Set("X-API-Key", readOnlyAPIKey)
		})

		require.Equal(t, http.StatusForbidden, meResp.StatusCode)
	})

	t.Run("Read-scoped API key can list projects", func(t *testing.T) {
		listResp := server.Client.GET("/api/v1/projects", func(r *http.Request) {
			r.Header.Set("X-API-Key", readOnlyAPIKey)
			r.Header.Set("X-Organization-ID", "global")
		})

		require.Equal(t, http.StatusOK, listResp.StatusCode)
	})

	t.Run("Read-scoped API key cannot create project", func(t *testing.T) {
		createProjectResp := server.Client.POST("/api/v1/projects", map[string]string{
			"name":   "blocked-project",
			"domain": "blocked.example.com",
		}, func(r *http.Request) {
			r.Header.Set("X-API-Key", readOnlyAPIKey)
			r.Header.Set("X-Organization-ID", "global")
		})

		require.Equal(t, http.StatusForbidden, createProjectResp.StatusCode)
	})

	t.Run("Manage-scoped API key can create project", func(t *testing.T) {
		createProjectResp := server.Client.POST("/api/v1/projects", map[string]string{
			"name":   "managed-project",
			"domain": "managed.example.com",
		}, func(r *http.Request) {
			r.Header.Set("X-API-Key", manageAPIKey)
			r.Header.Set("X-Organization-ID", "global")
		})

		require.Equal(t, http.StatusCreated, createProjectResp.StatusCode)
	})

	// 6. Revoke API Keys (Needs admin token again)
	server.Client.Token = loginData.Data.AccessToken
	revokeReadResp := server.Client.DELETE("/api/v1/api-keys/"+readOnlyAPIKeyID, func(r *http.Request) {
		r.Header.Set("X-Organization-ID", "global")
	})
	require.Equal(t, http.StatusOK, revokeReadResp.StatusCode)

	revokeManageResp := server.Client.DELETE("/api/v1/api-keys/"+manageAPIKeyID, func(r *http.Request) {
		r.Header.Set("X-Organization-ID", "global")
	})
	require.Equal(t, http.StatusOK, revokeManageResp.StatusCode)

	// 7. Verify revoked API Key no longer works
	server.Client.Token = ""
	failResp := server.Client.GET("/api/v1/projects", func(r *http.Request) {
		r.Header.Set("X-API-Key", readOnlyAPIKey)
		r.Header.Set("X-Organization-ID", "global")
	})
	assert.Equal(t, http.StatusUnauthorized, failResp.StatusCode)
}
