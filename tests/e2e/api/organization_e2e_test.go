//go:build e2e
// +build e2e

package api

import (
	"fmt"
	"testing"
	"time"

	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// Helper: Create user and login
func createUserAndLogin(t *testing.T, server *setup.TestServer) (string, *userEntity.User) {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("UserPass123!"), bcrypt.DefaultCost)

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	user := f.Create(func(u *userEntity.User) {
		u.Username = "user_org_" + uniqueSuffix
		u.Email = "org_" + uniqueSuffix + "@test.com"
		u.Password = string(hash)
	})

	// Login
	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": user.Username,
		"password": "UserPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	return loginRes.Data.AccessToken, user
}

func TestOrganizationE2E_Create_Success(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token, _ := createUserAndLogin(t, server)

	t.Run("Create Organization", func(t *testing.T) {
		payload := map[string]string{
			"name": "My Tech Startup",
			"slug": "my-tech-startup",
		}

		resp := server.Client.POST("/api/v1/organizations", payload, setup.WithAuth(token))

		if resp.StatusCode != 201 {
			var errRes map[string]interface{}
			_ = resp.JSON(&errRes)
			t.Logf("Response Body: %+v", errRes)
		}
		assert.Equal(t, 201, resp.StatusCode)

		var result struct {
			Data struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Slug    string `json:"slug"`
				OwnerID string `json:"owner_id"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Data.ID)
		assert.Equal(t, "My Tech Startup", result.Data.Name)
		assert.Equal(t, "my-tech-startup", result.Data.Slug)
	})

	t.Run("Create Duplicate Slug - Conflict", func(t *testing.T) {
		payload := map[string]string{
			"name": "Another Startup",
			"slug": "my-tech-startup", // Duplicate
		}

		resp := server.Client.POST("/api/v1/organizations", payload, setup.WithAuth(token))
		assert.Equal(t, 409, resp.StatusCode)
	})
}

func TestOrganizationE2E_CRUD(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token, _ := createUserAndLogin(t, server)

	// Prepare organization
	var orgID string
	var orgSlug = "crud-org-test"

	t.Run("1. Create", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/organizations", map[string]string{
			"name": "CRUD Test Org",
			"slug": orgSlug,
		}, setup.WithAuth(token))
		require.Equal(t, 201, resp.StatusCode)

		var result struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		resp.JSON(&result)
		orgID = result.Data.ID
	})

	t.Run("2. Get By ID", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/organizations/" + orgID, setup.WithAuth(token))
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("3. Get By Slug", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/organizations/slug/" + orgSlug, setup.WithAuth(token))
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("4. Update", func(t *testing.T) {
		resp := server.Client.PUT("/api/v1/organizations/" + orgID, map[string]string{
			"name": "Updated Org Name",
		}, setup.WithAuth(token))
		assert.Equal(t, 200, resp.StatusCode)

		// Verify update
		getResp := server.Client.GET("/api/v1/organizations/" + orgID, setup.WithAuth(token))
		var result struct {
			Data struct {
				Name string `json:"name"`
			} `json:"data"`
		}
		getResp.JSON(&result)
		assert.Equal(t, "Updated Org Name", result.Data.Name)
	})

	t.Run("5. List My Organizations", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/organizations/me", setup.WithAuth(token))
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data struct {
				Organizations []struct {
					ID string `json:"id"`
				} `json:"organizations"`
				Total int `json:"total"`
			} `json:"data"`
		}
		resp.JSON(&result)
		assert.GreaterOrEqual(t, result.Data.Total, 1)
	})

	t.Run("6. Delete", func(t *testing.T) {
		resp := server.Client.DELETE("/api/v1/organizations/" + orgID, setup.WithAuth(token))
		assert.Equal(t, 200, resp.StatusCode)

		// Verify retrieval fails
		getResp := server.Client.GET("/api/v1/organizations/" + orgID, setup.WithAuth(token))
		assert.Equal(t, 404, getResp.StatusCode)
	})
}
