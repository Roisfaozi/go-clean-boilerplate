//go:build e2e
// +build e2e

package api

import (
	"net/http"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createOrgWithToken creates an organization via the HTTP API and returns its ID.
func createOrgWithToken(t *testing.T, server *setup.TestServer, token, slug string) string {
	t.Helper()

	resp := server.Client.POST("/api/v1/organizations", map[string]string{
		"name": slug + " Org",
		"slug": slug,
	}, setup.WithAuth(token))
	require.Equal(t, 201, resp.StatusCode, "create org failed: %s", resp.String())

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, resp.JSON(&created))
	require.NotEmpty(t, created.Data.ID)
	return created.Data.ID
}

// createOrgRole creates a custom role in an organization via the HTTP API and returns its ID.
func createOrgRole(t *testing.T, server *setup.TestServer, token, orgID, name, description string) string {
	t.Helper()

	resp := server.Client.POST("/api/v1/organizations/"+orgID+"/roles", map[string]any{
		"name":        name,
		"description": description,
	}, setup.WithAuth(token))
	require.Equal(t, 201, resp.StatusCode, "create org role failed: %s", resp.String())

	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	require.NoError(t, resp.JSON(&created))
	require.NotEmpty(t, created.Data.ID)
	return created.Data.ID
}

func TestOrganizationRoleE2E_CrossTenantIsolation(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	// One user owns two organizations, so both tenant middleware (membership)
	// and Casbin pass for either org. Cross-tenant access must still fail at
	// the controller/role-scoping layer.
	token, _ := createUserAndLogin(t, server)
	suffix := uuid.New().String()[:8]
	orgA := createOrgWithToken(t, server, token, "e2e-cross-a-"+suffix)
	orgB := createOrgWithToken(t, server, token, "e2e-cross-b-"+suffix)

	roleAName := "e2e_editor_a"
	roleBName := "e2e_editor_b"
	roleBID := createOrgRole(t, server, token, orgB, roleBName, "Editor B")
	roleAID := createOrgRole(t, server, token, orgA, roleAName, "Editor A")

	t.Run("cross-tenant update returns 404 without disclosing role existence", func(t *testing.T) {
		resp := server.Client.PUT("/api/v1/organizations/"+orgA+"/roles/"+roleBID,
			map[string]any{"description": "hijacked"}, setup.WithAuth(token))
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("cross-tenant delete returns 404 without disclosing role existence", func(t *testing.T) {
		resp := server.Client.DELETE("/api/v1/organizations/"+orgA+"/roles/"+roleBID,
			setup.WithAuth(token))
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("listing never leaks roles across tenants", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/organizations/"+orgA+"/roles", setup.WithAuth(token))
		require.Equal(t, 200, resp.StatusCode)

		var out struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		require.NoError(t, resp.JSON(&out))
		for _, r := range out.Data {
			assert.NotEqual(t, roleBID, r.ID, "org B role must not appear in org A listing")
		}
	})

	t.Run("target tenant role survives cross-tenant attempts", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/organizations/"+orgB+"/roles", setup.WithAuth(token))
		require.Equal(t, 200, resp.StatusCode)

		var out struct {
			Data []struct {
				ID          string `json:"id"`
				Description string `json:"description"`
			} `json:"data"`
		}
		require.NoError(t, resp.JSON(&out))

		var found bool
		for _, r := range out.Data {
			if r.ID == roleBID {
				found = true
				assert.Equal(t, "Editor B", r.Description, "description must be untouched by cross-tenant attempt")
			}
		}
		assert.True(t, found, "org B role must still exist")
	})

	t.Run("non-member user is blocked from listing", func(t *testing.T) {
		outsiderToken, _ := createUserAndLogin(t, server)
		resp := server.Client.GET("/api/v1/organizations/"+orgA+"/roles", setup.WithAuth(outsiderToken))
		assert.Equal(t, 403, resp.StatusCode)
	})

	t.Run("own-tenant update and delete succeed", func(t *testing.T) {
		resp := server.Client.PUT("/api/v1/organizations/"+orgA+"/roles/"+roleAID,
			map[string]any{"description": "Updated A"}, setup.WithAuth(token))
		assert.Equal(t, 200, resp.StatusCode)

		resp = server.Client.DELETE("/api/v1/organizations/"+orgA+"/roles/"+roleAID,
			setup.WithAuth(token))
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestOrganizationRoleE2E_ApiKeyScopes(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token, _ := createUserAndLogin(t, server)
	orgID := createOrgWithToken(t, server, token, "e2e-scope-"+uuid.New().String()[:8])

	createKey := func(t *testing.T, scopes []string) string {
		t.Helper()
		resp := server.Client.POST("/api/v1/api-keys", map[string]any{
			"name":   "e2e-scope-key-" + uuid.New().String()[:8],
			"scopes": scopes,
		}, setup.WithAuth(token), setup.WithOrg(orgID))
		require.Equal(t, 201, resp.StatusCode, "create api key failed: %s", resp.String())

		var out struct {
			Data struct {
				Key string `json:"api_key"`
			} `json:"data"`
		}
		require.NoError(t, resp.JSON(&out))
		require.NotEmpty(t, out.Data.Key)
		return out.Data.Key
	}

	tests := []struct {
		name           string
		scopes         []string
		expectListCode int
	}{
		{
			// RequireScopeAuto derives org:view from the path AND the route
			// additionally requires role:view -> both are needed.
			name:           "org:view plus role:view can list",
			scopes:         []string{"org:view", "role:view"},
			expectListCode: http.StatusOK,
		},
		{
			name:           "role:view alone is blocked by derived org scope",
			scopes:         []string{"role:view"},
			expectListCode: http.StatusForbidden,
		},
		{
			name:           "org:view alone is blocked by explicit role scope",
			scopes:         []string{"org:view"},
			expectListCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := createKey(t, tt.scopes)
			resp := server.Client.GET("/api/v1/organizations/"+orgID+"/roles",
				func(r *http.Request) {
					r.Header.Set("X-API-Key", key)
				})
			assert.Equal(t, tt.expectListCode, resp.StatusCode,
				"unexpected status for scopes %v: %s", tt.scopes, resp.String())
		})
	}
}
