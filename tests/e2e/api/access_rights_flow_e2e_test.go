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

func TestAccessRightsFlowE2E(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	// 1. Setup Admin
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("AdminPass123!"), bcrypt.DefaultCost)
	admin := f.Create(func(u *userEntity.User) {
		u.Username = "flow_admin"
		u.Password = string(hash)
	})
	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	server.Enforcer.SavePolicy()

	resp := client.POST("/api/v1/auth/login", map[string]any{
		"username": admin.Username,
		"password": "AdminPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&loginRes)
	adminToken := loginRes.Data.AccessToken

	// 2. Create Access Right (Resource Group)
	resp = client.POST("/api/v1/access-rights", map[string]any{
		"name":        "User Management",
		"description": "Operations related to user accounts",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var arRes struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	resp.JSON(&arRes)
	arID := arRes.Data.ID

	// 3. Create Endpoint
	resp = client.POST("/api/v1/endpoints", map[string]any{
		"path":   "/api/v1/users",
		"method": "GET",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var epRes struct {
		Data struct {
			ID string `json:"id"`
		}
	}
	resp.JSON(&epRes)
	epID := epRes.Data.ID

	// 4. Link Endpoint to Access Right
	resp = client.POST("/api/v1/access-rights/link", map[string]any{
		"access_right_id": arID,
		"endpoint_id":     epID,
	}, setup.WithAuth(adminToken))
	assert.Equal(t, 200, resp.StatusCode)

	// 5. Create a specific Role
	roleID := uuid.New().String()
	server.DB.Create(&roleEntity.Role{ID: roleID, Name: "UserManager"})

	// 6. Grant Permission via the Endpoint details
	// (Simulating what the Access Matrix UI does)
	resp = client.POST("/api/v1/permissions/grant", map[string]any{
		"role":   "UserManager",
		"path":   "/api/v1/users",
		"method": "GET",
	}, setup.WithAuth(adminToken))
	assert.Equal(t, 201, resp.StatusCode)

	// 7. Verify Enforcement
	// Create a user with that role
	user := f.Create(func(u *userEntity.User) { u.Username = "manager_user"; u.Password = string(hash) })
	client.POST("/api/v1/permissions/assign-role", map[string]any{
		"user_id": user.ID,
		"role":    "UserManager",
	}, setup.WithAuth(adminToken))

	// Enforce check (sub, dom, obj, act)
	ok, err := server.Enforcer.Enforce(user.ID, "global", "/api/v1/users", "GET")
	require.NoError(t, err)
	assert.True(t, ok, "User should have access to linked endpoint")
}
