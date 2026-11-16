package test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	_ "github.com/glebarez/go-sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStatsTest(t *testing.T) (*sqlx.DB, usecase.StatsUseCase) {
	db, err := sqlx.Open("sqlite", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE users (id TEXT PRIMARY KEY, name TEXT, organization_id TEXT, deleted_at INTEGER DEFAULT 0);
		CREATE TABLE roles (id TEXT PRIMARY KEY, name TEXT, organization_id TEXT, deleted_at INTEGER DEFAULT 0);
		CREATE TABLE audit_logs (id TEXT PRIMARY KEY, action TEXT, organization_id TEXT, created_at INTEGER DEFAULT 0, deleted_at INTEGER DEFAULT 0);
		CREATE TABLE organization_members (id TEXT PRIMARY KEY, organization_id TEXT, user_id TEXT, role_id TEXT);
	`)
	require.NoError(t, err)

	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)

	uc := usecase.NewStatsUseCase(db, log)
	return db, uc
}

func TestStatsUseCase_GetDashboardSummary_Success(t *testing.T) {
	db, uc := setupStatsTest(t)

	// Seed test data
	db.MustExec(`INSERT INTO users (id, name, deleted_at) VALUES ('u1', 'Alice', 0)`)
	db.MustExec(`INSERT INTO users (id, name, deleted_at) VALUES ('u2', 'Bob', 0)`)
	db.MustExec(`INSERT INTO roles (id, name, deleted_at) VALUES ('r1', 'admin', 0)`)
	db.MustExec(`INSERT INTO audit_logs (id, action, created_at, deleted_at) VALUES ('a1', 'LOGIN', 1000, 0)`)
	db.MustExec(`INSERT INTO audit_logs (id, action, created_at, deleted_at) VALUES ('a2', 'LOGOUT', 2000, 0)`)
	db.MustExec(`INSERT INTO audit_logs (id, action, created_at, deleted_at) VALUES ('a3', 'LOGIN', 3000, 0)`)
	db.MustExec(`INSERT INTO organization_members (id) VALUES ('m1')`)
	db.MustExec(`INSERT INTO organization_members (id) VALUES ('m2')`)

	summary, err := uc.GetDashboardSummary(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, int64(2), summary.TotalUsers)
	assert.Equal(t, int64(1), summary.TotalRoles)
	assert.Equal(t, int64(3), summary.TotalAuditLogs)
	assert.Equal(t, int64(2), summary.TotalOrgMembers)
}

func TestStatsUseCase_GetDashboardSummary_EmptyDB(t *testing.T) {
	_, uc := setupStatsTest(t)

	summary, err := uc.GetDashboardSummary(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, int64(0), summary.TotalUsers)
	assert.Equal(t, int64(0), summary.TotalRoles)
	assert.Equal(t, int64(0), summary.TotalAuditLogs)
	assert.Equal(t, int64(0), summary.TotalOrgMembers)
}

func TestStatsUseCase_GetDashboardSummary_WithOrgScope(t *testing.T) {
	db, uc := setupStatsTest(t)

	// Seed data (without org scoping column, scope is a no-op on base context)
	db.MustExec(`INSERT INTO users (id, name, deleted_at) VALUES ('u1', 'Alice', 0)`)

	ctx := database.SetOrganizationContext(context.Background(), "org-1")
	summary, err := uc.GetDashboardSummary(ctx)

	// Should not error - scope is applied gracefully
	assert.NoError(t, err)
	assert.NotNil(t, summary)
}

func TestStatsUseCase_GetDashboardActivity_DefaultDays(t *testing.T) {
	_, uc := setupStatsTest(t)

	// days = 0 should default to 7
	activity, err := uc.GetDashboardActivity(context.Background(), 0)

	assert.NoError(t, err)
	assert.NotNil(t, activity)
	assert.Len(t, activity.Points, 7)
}

func TestStatsUseCase_GetDashboardActivity_NegativeDays(t *testing.T) {
	_, uc := setupStatsTest(t)

	// Negative days should default to 7
	activity, err := uc.GetDashboardActivity(context.Background(), -5)

	assert.NoError(t, err)
	assert.NotNil(t, activity)
	assert.Len(t, activity.Points, 7)
}

func TestStatsUseCase_GetDashboardActivity_CustomDays(t *testing.T) {
	_, uc := setupStatsTest(t)

	activity, err := uc.GetDashboardActivity(context.Background(), 14)

	assert.NoError(t, err)
	assert.NotNil(t, activity)
	assert.Len(t, activity.Points, 14)
}

func TestStatsUseCase_GetDashboardActivity_SingleDay(t *testing.T) {
	_, uc := setupStatsTest(t)

	activity, err := uc.GetDashboardActivity(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, activity)
	assert.Len(t, activity.Points, 1)
	assert.NotEmpty(t, activity.Points[0].Date)
}

func TestStatsUseCase_GetDashboardActivity_PointsHaveDates(t *testing.T) {
	_, uc := setupStatsTest(t)

	activity, err := uc.GetDashboardActivity(context.Background(), 3)

	assert.NoError(t, err)
	require.Len(t, activity.Points, 3)

	// All points should have non-empty dates
	for _, point := range activity.Points {
		assert.NotEmpty(t, point.Date)
		assert.GreaterOrEqual(t, point.Audits, int64(0))
		assert.GreaterOrEqual(t, point.Logins, int64(0))
	}
}

func TestStatsUseCase_GetDashboardActivity_PointsAreChronological(t *testing.T) {
	_, uc := setupStatsTest(t)

	activity, err := uc.GetDashboardActivity(context.Background(), 5)

	assert.NoError(t, err)
	require.Len(t, activity.Points, 5)

	// Points should be in chronological order (oldest first, newest last)
	for i := 1; i < len(activity.Points); i++ {
		assert.Less(t, activity.Points[i-1].Date, activity.Points[i].Date,
			"Points should be in chronological order")
	}
}

func TestStatsUseCase_GetSystemInsights_Success(t *testing.T) {
	_, uc := setupStatsTest(t)

	insights, err := uc.GetSystemInsights(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, insights)
	assert.IsType(t, &model.SystemInsights{}, insights)
	assert.NotEmpty(t, insights.MostActiveRole)
}
