//go:build e2e
// +build e2e

package modules

import (
	"net/http"
	"testing"

	apiKeyModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	userModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/model"
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

	// 4. Create API Key
	createReq := apiKeyModel.CreateApiKeyRequest{
		Name: "E2E Test Key",
	}
	createResp := server.Client.POST("/api/v1/api-keys", createReq, func(r *http.Request) {
		r.Header.Set("X-Organization-ID", "global")
	})
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createData struct {
		Data apiKeyModel.CreateApiKeyResponse `json:"data"`
	}
	_ = createResp.JSON(&createData)
	apiKey := createData.Data.Key
	apiKeyID := createData.Data.ID
	require.NotEmpty(t, apiKey)

	// 5. Use API Key to access /api/v1/users/me
	// Clear Bearer Token to ensure we use X-API-Key
	server.Client.Token = ""

	meResp := server.Client.GET("/api/v1/users/me", func(r *http.Request) {
		r.Header.Set("X-API-Key", apiKey)
	})

	require.Equal(t, http.StatusOK, meResp.StatusCode)

	var meData struct {
		Data userModel.UserResponse `json:"data"`
	}
	_ = meResp.JSON(&meData)
	assert.Equal(t, admin.ID, meData.Data.ID)
	assert.Equal(t, "api_admin", meData.Data.Username)

	// 6. Revoke API Key (Needs admin token again)
	server.Client.Token = loginData.Data.AccessToken
	revokeResp := server.Client.DELETE("/api/v1/api-keys/"+apiKeyID, func(r *http.Request) {
		r.Header.Set("X-Organization-ID", "global")
	})
	require.Equal(t, http.StatusOK, revokeResp.StatusCode)

	// 7. Verify API Key no longer works
	server.Client.Token = ""
	failResp := server.Client.GET("/api/v1/users/me", func(r *http.Request) {
		r.Header.Set("X-API-Key", apiKey)
	})
	assert.Equal(t, http.StatusUnauthorized, failResp.StatusCode)
}
