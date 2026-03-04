//go:build e2e
// +build e2e

package api

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/model"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPermissionE2E_AccessRightAssignment(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "ar_admin"
		u.Email = "ar@admin.com"
		u.Password = passHash
	})

	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	server.Enforcer.SavePolicy()

	loginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&loginRes)
	adminToken := loginRes.Data.AccessToken

	// Create Access Right and Endpoints via DB
	ep1 := entity.Endpoint{Path: "/api/test", Method: "GET"}
	ep2 := entity.Endpoint{Path: "/api/test", Method: "POST"}
	server.DB.Create(&ep1)
	server.DB.Create(&ep2)

	ar := entity.AccessRight{
		Name:      "Test Access Right",
		Endpoints: []entity.Endpoint{ep1, ep2},
	}
	server.DB.Create(&ar)

	roleName := "TestRole"

	t.Run("GetRoleAccessRights - Unassigned", func(t *testing.T) {
		resp := client.GET("/api/v1/permissions/roles/"+roleName+"/access-rights", setup.WithAuth(adminToken))
		require.Equal(t, 200, resp.StatusCode)

		var res struct {
			Data []model.RoleAccessRightStatus `json:"data"`
		}
		err := resp.JSON(&res)
		require.NoError(t, err)

		found := false
		for _, s := range res.Data {
			if s.ID == ar.ID {
				found = true
				assert.False(t, s.Assigned)
				assert.False(t, s.Partial)
			}
		}
		assert.True(t, found, "Access right should be in the list")
	})

	t.Run("AssignAccessRight", func(t *testing.T) {
		payload := map[string]any{
			"role":            roleName,
			"access_right_id": ar.ID,
		}
		resp := client.POST("/api/v1/permissions/assign-access-right", payload, setup.WithAuth(adminToken))
		require.Equal(t, 200, resp.StatusCode)

		ok, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "GET")
		assert.True(t, ok)
		ok2, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "POST")
		assert.True(t, ok2)
	})

	t.Run("GetRoleAccessRights - Assigned", func(t *testing.T) {
		resp := client.GET("/api/v1/permissions/roles/"+roleName+"/access-rights", setup.WithAuth(adminToken))
		require.Equal(t, 200, resp.StatusCode)

		var res struct {
			Data []model.RoleAccessRightStatus `json:"data"`
		}
		resp.JSON(&res)

		for _, s := range res.Data {
			if s.ID == ar.ID {
				assert.True(t, s.Assigned)
			}
		}
	})

	t.Run("RevokeAccessRight", func(t *testing.T) {
		payload := map[string]any{
			"role":            roleName,
			"access_right_id": ar.ID,
		}
		resp := client.DELETE("/api/v1/permissions/revoke-access-right", payload, setup.WithAuth(adminToken))
		require.Equal(t, 200, resp.StatusCode)

		ok, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "GET")
		assert.False(t, ok)
		ok2, _ := server.Enforcer.Enforce(roleName, "global", "/api/test", "POST")
		assert.False(t, ok2)
	})
}
