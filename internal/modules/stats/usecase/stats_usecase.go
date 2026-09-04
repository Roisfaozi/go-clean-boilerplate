package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type StatsUseCase interface {
	GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error)
	GetDashboardActivity(ctx context.Context, days int) (*model.DashboardActivity, error)
	GetSystemInsights(ctx context.Context) (*model.SystemInsights, error)
}

type statsUseCase struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewStatsUseCase(db any, log *logrus.Logger) StatsUseCase {
	return &statsUseCase{
		db:  tx.ExtractSQLX(db),
		log: log,
	}
}

func (u *statsUseCase) getDB(ctx context.Context) tx.DBTX {
	if dbtx, ok := tx.DBTXFromContext(ctx); ok {
		return dbtx
	}
	return u.db
}

func (u *statsUseCase) countWithOrg(ctx context.Context, table string, col string) int64 {
	whereClauses := []string{fmt.Sprintf("%s.deleted_at = 0", table)}
	var args []any

	orgClause, orgArgs := database.SQLOrganizationClause(ctx, fmt.Sprintf("%s.%s", table, col))
	if orgClause != "" {
		whereClauses = append(whereClauses, orgClause)
		args = append(args, orgArgs...)
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, table, strings.Join(whereClauses, " AND "))
	var count int64
	_ = u.getDB(ctx).GetContext(ctx, &count, query, args...)
	return count
}

func (u *statsUseCase) GetDashboardSummary(ctx context.Context) (*model.DashboardSummary, error) {
	var summary model.DashboardSummary

	orgID := database.GetOrganizationID(ctx)
	if orgID != "" {
		var totalUsers int64
		query := `SELECT COUNT(DISTINCT user_id) FROM organization_members WHERE organization_id = ?`
		_ = u.getDB(ctx).GetContext(ctx, &totalUsers, query, orgID)
		summary.TotalUsers = totalUsers
	} else {
		summary.TotalUsers = u.countWithOrg(ctx, "users", "organization_id")
	}

	summary.TotalRoles = u.countWithOrg(ctx, "roles", "organization_id")
	summary.TotalAuditLogs = u.countWithOrg(ctx, "audit_logs", "organization_id")

	// Organization members count
	whereClauses := []string{"1=1"}
	var orgArgs []any
	if orgID != "" {
		whereClauses = append(whereClauses, "organization_id = ?")
		orgArgs = append(orgArgs, orgID)
	}
	queryMem := fmt.Sprintf(`SELECT COUNT(*) FROM organization_members WHERE %s`, strings.Join(whereClauses, " AND "))
	var totalOrgMembers int64
	_ = u.getDB(ctx).GetContext(ctx, &totalOrgMembers, queryMem, orgArgs...)
	summary.TotalOrgMembers = totalOrgMembers

	return &summary, nil
}

func (u *statsUseCase) GetDashboardActivity(ctx context.Context, days int) (*model.DashboardActivity, error) {
	if days <= 0 {
		days = 7
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.UTC
	}

	points := make([]model.ActivityPoint, days)
	now := time.Now().In(loc)

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc).UnixMilli()
		endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999, loc).UnixMilli()

		whereClauses := []string{"deleted_at = 0", "created_at >= ?", "created_at <= ?"}
		args := []any{startOfDay, endOfDay}

		orgClause, orgArgs := database.SQLOrganizationClause(ctx, "organization_id")
		if orgClause != "" {
			whereClauses = append(whereClauses, orgClause)
			args = append(args, orgArgs...)
		}

		whereSQL := strings.Join(whereClauses, " AND ")

		var auditCount int64
		qAudit := fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs WHERE %s`, whereSQL)
		_ = u.getDB(ctx).GetContext(ctx, &auditCount, qAudit, args...)

		var loginCount int64
		loginArgs := append([]any{"LOGIN"}, args...)
		qLogin := fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs WHERE action = ? AND %s`, whereSQL)
		_ = u.getDB(ctx).GetContext(ctx, &loginCount, qLogin, loginArgs...)

		points[days-1-i] = model.ActivityPoint{
			Date:   dateStr,
			Audits: auditCount,
			Logins: loginCount,
		}
	}

	return &model.DashboardActivity{Points: points}, nil
}

func (u *statsUseCase) GetSystemInsights(ctx context.Context) (*model.SystemInsights, error) {
	whereClauses := []string{"1=1"}
	var args []any

	if orgID := database.GetOrganizationID(ctx); orgID != "" {
		whereClauses = append(whereClauses, "om.organization_id = ?")
		args = append(args, orgID)
	}

	query := fmt.Sprintf(`
		SELECT r.name
		FROM organization_members om
		INNER JOIN roles r ON r.id = om.role_id
		WHERE %s
		GROUP BY r.name
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`, strings.Join(whereClauses, " AND "))

	var roleName string
	mostActive := "none"
	err := u.getDB(ctx).GetContext(ctx, &roleName, query, args...)
	if err == nil && roleName != "" {
		mostActive = roleName
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		u.log.WithError(err).Warn("Failed to fetch most active role insight")
	}

	return &model.SystemInsights{
		MostActiveRole: mostActive,
	}, nil
}
