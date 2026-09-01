package usecase

import (
	"context"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StatsUseCase interface {
	GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error)
	GetDashboardActivity(ctx context.Context, days int) (*model.DashboardActivity, error)
	GetSystemInsights(ctx context.Context) (*model.SystemInsights, error)
}

type statsUseCase struct {
	db  *gorm.DB
	log *logrus.Logger
}

func NewStatsUseCase(db *gorm.DB, log *logrus.Logger) StatsUseCase {
	return &statsUseCase{
		db:  db,
		log: log,
	}
}

func (u *statsUseCase) GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error) {
	var summary model.DashboardSummary

	// Apply organization scope if present
	scope := database.OrganizationScope(ctx)
	// For tenant user count, prefer counting active memberships so global users are
	// not misattributed. If organization context exists, count distinct user IDs
	// in organization_members. Otherwise fallback to users table for global count.
	orgID := database.GetOrganizationID(ctx)
	if orgID != "" {
		var totalUsers int64
		u.db.WithContext(ctx).Table("organization_members").Where("organization_id = ?", orgID).Distinct("user_id").Count(&totalUsers)
		summary.TotalUsers = totalUsers
	} else {
		u.db.WithContext(ctx).Model(&struct{ Table string }{}).Table("users").Scopes(scope).Count(&summary.TotalUsers)
	}
	u.db.WithContext(ctx).Model(&struct{ Table string }{}).Table("roles").Scopes(scope).Count(&summary.TotalRoles)
	u.db.WithContext(ctx).Model(&struct{ Table string }{}).Table("audit_logs").Scopes(scope).Count(&summary.TotalAuditLogs)
	u.db.WithContext(ctx).Model(&struct{ Table string }{}).Table("organization_members").Scopes(scope).Count(&summary.TotalOrgMembers)

	return &summary, nil
}

func (u *statsUseCase) GetDashboardActivity(ctx context.Context, days int) (*model.DashboardActivity, error) {
	if days <= 0 {
		days = 7
	}

	// TODO: read timezone from clinic/branch settings once setting model exposes branch_timezone
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}

	scope := database.OrganizationScope(ctx)
	points := make([]model.ActivityPoint, days)
	now := time.Now().In(loc)

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc).UnixMilli()
		endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999, loc).UnixMilli()

		var auditCount, loginCount int64

		// Count all audits for that day
		u.db.WithContext(ctx).Table("audit_logs").
			Scopes(scope).
			Where("created_at >= ? AND created_at <= ?", startOfDay, endOfDay).
			Count(&auditCount)

		// Count only LOGIN actions
		u.db.WithContext(ctx).Table("audit_logs").
			Scopes(scope).
			Where("action = ? AND created_at >= ? AND created_at <= ?", "LOGIN", startOfDay, endOfDay).
			Count(&loginCount)

		points[days-1-i] = model.ActivityPoint{
			Date:   dateStr,
			Audits: auditCount,
			Logins: loginCount,
		}
	}

	return &model.DashboardActivity{Points: points}, nil
}

func (u *statsUseCase) GetSystemInsights(ctx context.Context) (*model.SystemInsights, error) {
	var result struct {
		Role string
	}

	query := u.db.WithContext(ctx).Table("organization_members om").
		Select("r.name as role, COUNT(*) as total").
		Joins("JOIN roles r ON r.id = om.role_id").
		Group("r.name").
		Order("total DESC").
		Limit(1)

	if orgID := database.GetOrganizationID(ctx); orgID != "" {
		query = query.Where("om.organization_id = ?", orgID)
	}

	mostActive := "none"
	if err := query.Scan(&result).Error; err == nil && result.Role != "" {
		mostActive = result.Role
	}

	return &model.SystemInsights{
		MostActiveRole: mostActive,
	}, nil
}
