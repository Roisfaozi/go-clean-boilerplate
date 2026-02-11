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

func setupProjectE2E(t *testing.T, server *setup.TestServer) (token string, orgID string) {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("ProjPass123!"), bcrypt.DefaultCost)

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	f.Create(func(u *userEntity.User) {
		u.Username = "proj_user_" + uniqueSuffix
		u.Email = "proj_" + uniqueSuffix + "@test.com"
		u.Password = string(hash)
	})

	// Login
	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": "proj_user_" + uniqueSuffix,
		"password": "ProjPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	token = loginRes.Data.AccessToken

	// Create an organization (required for tenant-scoped project routes)
	orgResp := server.Client.POST("/api/v1/organizations", map[string]string{
		"name": "Project Test Org " + uniqueSuffix,
		"slug": "project-org-" + uniqueSuffix,
	}, setup.WithAuth(token))
	require.Equal(t, 201, orgResp.StatusCode)

	var orgResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	orgResp.JSON(&orgResult)
	orgID = orgResult.Data.ID

	return token, orgID
}

func TestProjectE2E_CRUD_Lifecycle(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token, orgID := setupProjectE2E(t, server)

	var projectID string

	t.Run("1. Create Project", func(t *testing.T) {
		payload := map[string]string{
			"name":   "Test Project Alpha",
			"domain": "alpha.example.com",
		}

		resp := server.Client.POST("/api/v1/projects", payload,
			setup.WithAuth(token),
			setup.WithOrg(orgID),
		)

		if resp.StatusCode != 201 {
			t.Logf("Create Response: %s", resp.String())
		}
		require.Equal(t, 201, resp.StatusCode)

		var result struct {
			Data struct {
				ID             string `json:"id"`
				Name           string `json:"name"`
				Domain         string `json:"domain"`
				Status         string `json:"status"`
				OrganizationID string `json:"organization_id"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Data.ID)
		assert.Equal(t, "Test Project Alpha", result.Data.Name)
		assert.Equal(t, "alpha.example.com", result.Data.Domain)
		assert.Equal(t, "active", result.Data.Status)
		projectID = result.Data.ID
	})

	t.Run("2. Get All Projects", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/projects",
			setup.WithAuth(token),
			setup.WithOrg(orgID),
		)
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"data"`
		}
		resp.JSON(&result)
		assert.GreaterOrEqual(t, len(result.Data), 1)
	})

	t.Run("3. Get Project By ID", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/projects/"+projectID,
			setup.WithAuth(token),
			setup.WithOrg(orgID),
		)
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Domain string `json:"domain"`
			} `json:"data"`
		}
		resp.JSON(&result)
		assert.Equal(t, projectID, result.Data.ID)
		assert.Equal(t, "Test Project Alpha", result.Data.Name)
	})

	t.Run("4. Update Project", func(t *testing.T) {
		payload := map[string]string{
			"name":   "Updated Project Alpha",
			"domain": "updated-alpha.example.com",
		}
		resp := server.Client.PUT("/api/v1/projects/"+projectID, payload,
			setup.WithAuth(token),
			setup.WithOrg(orgID),
		)
		assert.Equal(t, 200, resp.StatusCode)

		// Verify the update
		getResp := server.Client.GET("/api/v1/projects/"+projectID,
			setup.WithAuth(token),
			setup.WithOrg(orgID),
		)
		var result struct {
			Data struct {
				Name   string `json:"name"`
				Domain string `json:"domain"`
			} `json:"data"`
		}
		getResp.JSON(&result)
		assert.Equal(t, "Updated Project Alpha", result.Data.Name)
		assert.Equal(t, "updated-alpha.example.com", result.Data.Domain)
	})

	t.Run("5. Delete Project", func(t *testing.T) {
		resp := server.Client.DELETE("/api/v1/projects/"+projectID,
			setup.WithAuth(token),
			setup.WithOrg(orgID),
		)
		assert.Equal(t, 200, resp.StatusCode)

		// Verify deletion
		getResp := server.Client.GET("/api/v1/projects/"+projectID,
			setup.WithAuth(token),
			setup.WithOrg(orgID),
		)
		assert.Equal(t, 404, getResp.StatusCode)
	})
}

func TestProjectE2E_CreateMultiple(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token, orgID := setupProjectE2E(t, server)

	// Create 3 projects
	for i := 1; i <= 3; i++ {
		payload := map[string]string{
			"name":   fmt.Sprintf("Project %d", i),
			"domain": fmt.Sprintf("project%d.example.com", i),
		}
		resp := server.Client.POST("/api/v1/projects", payload,
			setup.WithAuth(token),
			setup.WithOrg(orgID),
		)
		require.Equal(t, 201, resp.StatusCode)
	}

	// List all projects
	resp := server.Client.GET("/api/v1/projects",
		setup.WithAuth(token),
		setup.WithOrg(orgID),
	)
	assert.Equal(t, 200, resp.StatusCode)

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&result)
	assert.GreaterOrEqual(t, len(result.Data), 3)
}

func TestProjectE2E_Unauthorized(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	t.Run("Create Without Auth", func(t *testing.T) {
		resp := server.Client.POST("/api/v1/projects", map[string]string{
			"name":   "Unauth Project",
			"domain": "unauth.com",
		})
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Get All Without Auth", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/projects")
		assert.Equal(t, 401, resp.StatusCode)
	})
}

func TestProjectE2E_GetByID_NotFound(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token, orgID := setupProjectE2E(t, server)

	resp := server.Client.GET("/api/v1/projects/nonexistent-id",
		setup.WithAuth(token),
		setup.WithOrg(orgID),
	)
	assert.Equal(t, 404, resp.StatusCode)
}
