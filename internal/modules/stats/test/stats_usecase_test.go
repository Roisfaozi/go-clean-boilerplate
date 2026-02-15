package test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testTable represents tables we create for stats counting
type testUser struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

func (testUser) TableName() string { return "users" }

type testRole struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

func (testRole) TableName() string { return "roles" }

type testAuditLog struct {
	ID        string `gorm:"primaryKey"`
	Action    string
	CreatedAt int64
}

func (testAuditLog) TableName() string { return "audit_logs" }

type testOrgMember struct {
	ID string `gorm:"primaryKey"`
}

func (testOrgMember) TableName() string { return "organization_members" }

func setupStatsTest(t *testing.T) (*gorm.DB, usecase.StatsUseCase) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Migrate test tables
	err = db.AutoMigrate(&testUser{}, &testRole{}, &testAuditLog{}, &testOrgMember{})
	require.NoError(t, err)

	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)

	uc := usecase.NewStatsUseCase(db, log)
	return db, uc
}

func TestStatsUseCase_GetDashboardSummary_Success(t *testing.T) {
	db, uc := setupStatsTest(t)

	// Seed test data
	db.Create(&testUser{ID: "u1", Name: "Alice"})
	db.Create(&testUser{ID: "u2", Name: "Bob"})
	db.Create(&testRole{ID: "r1", Name: "admin"})
	db.Create(&testAuditLog{ID: "a1", Action: "LOGIN", CreatedAt: 1000})
	db.Create(&testAuditLog{ID: "a2", Action: "LOGOUT", CreatedAt: 2000})
	db.Create(&testAuditLog{ID: "a3", Action: "LOGIN", CreatedAt: 3000})
	db.Create(&testOrgMember{ID: "m1"})
	db.Create(&testOrgMember{ID: "m2"})

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
	db.Create(&testUser{ID: "u1", Name: "Alice"})

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
	assert.Greater(t, insights.AvgLatencyMs, float64(0))
	assert.GreaterOrEqual(t, insights.ErrorRate, float64(0))
	assert.NotEmpty(t, insights.Uptime)
	assert.NotEmpty(t, insights.MostActiveRole)
}

func TestStatsUseCase_GetSystemInsights_ReturnsExpectedStructure(t *testing.T) {
	_, uc := setupStatsTest(t)

	insights, err := uc.GetSystemInsights(context.Background())

	assert.NoError(t, err)
	assert.IsType(t, &model.SystemInsights{}, insights)
	assert.Equal(t, 24.5, insights.AvgLatencyMs)
	assert.Equal(t, 0.02, insights.ErrorRate)
	assert.Equal(t, "99.99%", insights.Uptime)
	assert.Equal(t, "role:admin", insights.MostActiveRole)
}
