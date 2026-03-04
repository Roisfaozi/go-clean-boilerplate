//go:build e2e
// +build e2e

package api

import (
	"testing"
	"time"

	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestSearchAuditE2E_DynamicSearchAndAudit(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "audit_admin"
		u.Email = "audit@admin.com"
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

	targetUser := f.Create(func(u *userEntity.User) {
		u.Username = "target_update"
		u.Email = "target@update.com"
		u.Password = passHash
		u.Name = "Old Name"
	})

	updatePayload := map[string]any{
		"name":     "New Name Updated",
		"username": targetUser.Username,
	}

	resp = client.POST("/api/v1/auth/login", map[string]any{"username": targetUser.Username, "password": "StrongPass123!"})
	var targetLoginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&targetLoginRes)
	targetToken := targetLoginRes.Data.AccessToken

	resp = client.PUT("/api/v1/users/me", updatePayload, setup.WithAuth(targetToken))
	require.Equal(t, 200, resp.StatusCode)

	// Wait for audit log async processing (Outbox -> Worker -> Log)
	// Outbox worker runs every 5s, we wait 10s to be safe
	time.Sleep(10 * time.Second)

	searchPayload := map[string]any{
		"filter": map[string]any{
			"action":    map[string]any{"type": "equals", "from": "UPDATE"},
			"entity":    map[string]any{"type": "equals", "from": "User"},
			"entity_id": map[string]any{"type": "equals", "from": targetUser.ID},
		},
		"page":      1,
		"page_size": 10,
	}

	resp = client.POST("/api/v1/audit-logs/search", searchPayload, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	var searchRes struct {
		Data []struct {
			ID        string `json:"id"`
			Action    string `json:"action"`
			Entity    string `json:"entity"`
			OldValues any    `json:"old_values"`
			NewValues any    `json:"new_values"`
		} `json:"data"`
		Paging struct {
			Total int64 `json:"total"`
		} `json:"paging"`
	}
	err := resp.JSON(&searchRes)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, searchRes.Paging.Total, int64(1))
	if len(searchRes.Data) > 0 {
		log := searchRes.Data[0]
		assert.Equal(t, "UPDATE", log.Action)
		assert.Equal(t, "User", log.Entity)
	}
}

func TestSearchAuditE2E_UserDynamicSearch(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "search_admin"
		u.Email = "search@admin.com"
		u.Password = passHash
	})
	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	server.Enforcer.SavePolicy()

	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		}
	}
	resp.JSON(&loginRes)
	token := loginRes.Data.AccessToken

	f.Create(func(u *userEntity.User) {
		u.Name = "Alice Wonderland"
		u.Email = "alice@test.com"
		u.Username = "alice_w"
	})
	f.Create(func(u *userEntity.User) { u.Name = "Bob Builder"; u.Email = "bob@test.com"; u.Username = "bob_b" })
	f.Create(func(u *userEntity.User) {
		u.Name = "Charlie Chocolate"
		u.Email = "charlie@test.com"
		u.Username = "charlie_c"
	})

	searchPayload := map[string]any{
		"filter": map[string]any{
			"email": map[string]any{"type": "contains", "from": "alice"},
		},
	}
	resp = client.POST("/api/v1/users/search", searchPayload, setup.WithAuth(token))
	require.Equal(t, 200, resp.StatusCode)

	var userSearchRes struct {
		Data []struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"data"`
		Paging struct {
			Total int64 `json:"total"`
		} `json:"paging"`
	}
	resp.JSON(&userSearchRes)

	if userSearchRes.Paging.Total == 0 {
		t.Logf("DEBUG: Search Result Empty. Payload: %+v. Response: %+v", searchPayload, userSearchRes)
	}

	assert.Equal(t, int64(1), userSearchRes.Paging.Total)
	if len(userSearchRes.Data) > 0 {
		assert.Equal(t, "alice@test.com", userSearchRes.Data[0].Email)
	}

	sortPayload := map[string]any{
		"sort": []map[string]any{
			{"colId": "name", "sort": "desc"},
		},
		"page_size": 5,
	}
	resp = client.POST("/api/v1/users/search", sortPayload, setup.WithAuth(token))
	require.Equal(t, 200, resp.StatusCode)

	var sortedRes struct {
		Data []struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	resp.JSON(&sortedRes)

	assert.NotEmpty(t, sortedRes.Data)
}
