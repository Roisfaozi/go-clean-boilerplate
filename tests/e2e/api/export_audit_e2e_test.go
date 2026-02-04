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

func TestAuditExportE2E(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()
	client := server.Client

	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	passHash := string(hash)

	// Create Admin
	admin := f.Create(func(u *userEntity.User) {
		u.Username = "export_admin"
		u.Email = "export@admin.com"
		u.Password = passHash
	})
	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	server.Enforcer.SavePolicy()

	// Login Admin
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

	// Generate Audit Log (Update User)
	targetUser := f.Create(func(u *userEntity.User) {
		u.Username = "target_export"
		u.Email = "target@export.com"
		u.Password = passHash
		u.Name = "Original Name"
	})

	updatePayload := map[string]any{
		"name":     "Updated Name Export",
		"username": targetUser.Username,
	}

	// Login as target to update self
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

	// Wait for audit log async processing if any
	time.Sleep(2 * time.Second)

	// Test Export
	// Get today's date (UTC to match server handling)
	today := time.Now().UTC().Format("2006-01-02")

	resp = client.GET("/api/v1/audit-logs/export?from_date="+today+"&to_date="+today, setup.WithAuth(adminToken))
	require.Equal(t, 200, resp.StatusCode)

	// Verify Headers
	assert.Equal(t, "text/csv", resp.Header.Get("Content-Type"))
	assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment; filename=audit_logs.csv")

	// Verify Body
	body := resp.String()
	// Check Header row
	assert.Contains(t, body, "ID,UserID,Action,Entity,EntityID,OldValues,NewValues,IPAddress,UserAgent,CreatedAt")
	// Check Data row
	assert.Contains(t, body, "UPDATE")
	assert.Contains(t, body, "User")
	assert.Contains(t, body, targetUser.ID)
	assert.Contains(t, body, "Updated Name Export")
}
