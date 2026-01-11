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

	// 1. Setup: Create Admin User (to search audits)
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "audit_admin"
		u.Email = "audit@admin.com"
		u.Password = passHash
	})
	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin")
	server.Enforcer.AddPolicy("role:superadmin", "*", "*") 
	server.Enforcer.SavePolicy()

	// Login Admin
	loginPayload := map[string]any{"username": admin.Username, "password": "StrongPass123!"}
	resp := client.POST("/api/v1/auth/login", loginPayload)
	require.Equal(t, 200, resp.StatusCode)
	
	var loginRes struct { Data struct { AccessToken string `json:"access_token"` } }
	resp.JSON(&loginRes)
	adminToken := loginRes.Data.AccessToken

	// 2. Perform Action: Update User (Generates Audit Log)
	// Create another user to update
	targetUser := f.Create(func(u *userEntity.User) {
		u.Username = "target_update"
		u.Email = "target@update.com"
		u.Password = passHash
		u.Name = "Old Name"
	})

	updatePayload := map[string]any{
		"name":     "New Name Updated",
		"username": targetUser.Username, // Required by validation
	}
	
	// Login Target User
	resp = client.POST("/api/v1/auth/login", map[string]any{"username": targetUser.Username, "password": "StrongPass123!"})
	var targetLoginRes struct { Data struct { AccessToken string `json:"access_token"` } }
	resp.JSON(&targetLoginRes)
	targetToken := targetLoginRes.Data.AccessToken

	// Update Self
	resp = client.PUT("/api/v1/users/me", updatePayload, setup.WithAuth(targetToken))
	require.Equal(t, 200, resp.StatusCode)

	// Allow slight delay for async audit logging (if async)
	time.Sleep(2 * time.Second)

	// 3. Dynamic Search on Audit Logs (Admin Only)
	// Use Struct Field Names for filtering to ensure mapping works with QueryBuilder's GetDBFieldName
	searchPayload := map[string]any{
		"filter": map[string]any{
			"Action":   map[string]any{"type": "equals", "from": "UPDATE"},
			"Entity":   map[string]any{"type": "equals", "from": "User"},
			"EntityID": map[string]any{"type": "equals", "from": targetUser.ID},
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
		Meta struct {
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	err := resp.JSON(&searchRes)
	require.NoError(t, err)

	// Assertions
	assert.GreaterOrEqual(t, searchRes.Meta.Total, int64(1))
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

	// 1. Setup Admin
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "search_admin"
		u.Email = "search@admin.com"
		u.Password = passHash
	})
	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin")
	server.Enforcer.AddPolicy("role:superadmin", "*", "*")
	server.Enforcer.SavePolicy()

	// Login
	resp := client.POST("/api/v1/auth/login", map[string]any{"username": admin.Username, "password": "StrongPass123!"})
	var loginRes struct { Data struct { AccessToken string `json:"access_token"` } }
	resp.JSON(&loginRes)
	token := loginRes.Data.AccessToken

	// 2. Create Dummy Users for Searching
	f.Create(func(u *userEntity.User) { u.Name = "Alice Wonderland"; u.Email = "alice@test.com"; u.Username = "alice_w" })
	f.Create(func(u *userEntity.User) { u.Name = "Bob Builder"; u.Email = "bob@test.com"; u.Username = "bob_b" })
	f.Create(func(u *userEntity.User) { u.Name = "Charlie Chocolate"; u.Email = "charlie@test.com"; u.Username = "charlie_c" })

	// 3. Search for "Alice" using Struct Field Name "Email"
	searchPayload := map[string]any{
		"filter": map[string]any{
			"Email": map[string]any{"type": "contains", "from": "alice"},
		},
	}
	resp = client.POST("/api/v1/users/search", searchPayload, setup.WithAuth(token))
	require.Equal(t, 200, resp.StatusCode)

	var userSearchRes struct {
		Data []struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"data"`
		Meta struct {
			Total int64 `json:"total"`
		} `json:"meta"`
	}
	resp.JSON(&userSearchRes)
	
	// Debug log
	if userSearchRes.Meta.Total == 0 {
		t.Logf("DEBUG: Search Result Empty. Payload: %+v. Response: %+v", searchPayload, userSearchRes)
	}

	assert.Equal(t, int64(1), userSearchRes.Meta.Total)
	if len(userSearchRes.Data) > 0 {
		assert.Equal(t, "alice@test.com", userSearchRes.Data[0].Email)
	}

	// 4. Complex Search (Sorting)
	sortPayload := map[string]any{
		"sort": []map[string]any{
			{"colId": "Name", "sort": "desc"}, // Use "Name" instead of "name" just to be safe
		},
		"page_size": 5,
	}
	resp = client.POST("/api/v1/users/search", sortPayload, setup.WithAuth(token))
	require.Equal(t, 200, resp.StatusCode)
	
	var sortedRes struct { Data []struct { Name string `json:"name"` } `json:"data"` }
	resp.JSON(&sortedRes)
	
	assert.NotEmpty(t, sortedRes.Data)
}
