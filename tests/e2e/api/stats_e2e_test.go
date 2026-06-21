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

func loginForStats(t *testing.T, server *setup.TestServer) string {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("StatsPass123!"), bcrypt.DefaultCost)

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	f.Create(func(u *userEntity.User) {
		u.Username = "stats_user_" + uniqueSuffix
		u.Email = "stats_" + uniqueSuffix + "@test.com"
		u.Password = string(hash)
	})

	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": "stats_user_" + uniqueSuffix,
		"password": "StatsPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	return loginRes.Data.AccessToken
}

func TestStatsE2E_GetSummary(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token := loginForStats(t, server)

	resp := server.Client.GET("/api/v1/stats/summary", setup.WithAuth(token))

	assert.Equal(t, 200, resp.StatusCode)

	var result struct {
		Data struct {
			TotalUsers      int64 `json:"total_users"`
			TotalRoles      int64 `json:"total_roles"`
			TotalAuditLogs  int64 `json:"total_audit_logs"`
			TotalOrgMembers int64 `json:"total_org_members"`
		} `json:"data"`
	}
	err := resp.JSON(&result)
	require.NoError(t, err)

	// At least the test user we created should exist
	assert.GreaterOrEqual(t, result.Data.TotalUsers, int64(1))
}

func TestStatsE2E_GetActivity(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token := loginForStats(t, server)

	t.Run("Default Days", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/stats/activity", setup.WithAuth(token))

		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data struct {
				Points []struct {
					Date   string `json:"date"`
					Audits int64  `json:"audits"`
					Logins int64  `json:"logins"`
				} `json:"points"`
			} `json:"data"`
		}
		err := resp.JSON(&result)
		require.NoError(t, err)
		assert.Len(t, result.Data.Points, 7, "Default should return 7 days of activity")
	})

	t.Run("Custom Days", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/stats/activity?days=14", setup.WithAuth(token))
		assert.Equal(t, 200, resp.StatusCode)

		var result struct {
			Data struct {
				Points []struct {
					Date string `json:"date"`
				} `json:"points"`
			} `json:"data"`
		}
		resp.JSON(&result)
		assert.Len(t, result.Data.Points, 14, "Should return 14 days of activity")
	})
}

func TestStatsE2E_GetInsights(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token := loginForStats(t, server)

	resp := server.Client.GET("/api/v1/stats/insights", setup.WithAuth(token))

	assert.Equal(t, 200, resp.StatusCode)

	var result struct {
		Data struct {
			AvgLatencyMs   float64 `json:"avg_latency_ms"`
			ErrorRate      float64 `json:"error_rate"`
			Uptime         string  `json:"uptime"`
			MostActiveRole string  `json:"most_active_role"`
		} `json:"data"`
	}
	err := resp.JSON(&result)
	require.NoError(t, err)
	assert.Greater(t, result.Data.AvgLatencyMs, float64(0))
	assert.NotEmpty(t, result.Data.Uptime)
	assert.NotEmpty(t, result.Data.MostActiveRole)
}

func TestStatsE2E_Unauthorized(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	t.Run("Summary Without Auth", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/stats/summary")
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Activity Without Auth", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/stats/activity")
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Insights Without Auth", func(t *testing.T) {
		resp := server.Client.GET("/api/v1/stats/insights")
		assert.Equal(t, 401, resp.StatusCode)
	})
}
