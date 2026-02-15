package model

type DashboardSummary struct {
	TotalUsers      int64 `json:"total_users"`
	TotalRoles      int64 `json:"total_roles"`
	TotalAuditLogs  int64 `json:"total_audit_logs"`
	TotalOrgMembers int64 `json:"total_org_members"`
}

type ActivityPoint struct {
	Date   string `json:"date"`
	Audits int64  `json:"audits"`
	Logins int64  `json:"logins"`
}

type DashboardActivity struct {
	Points []ActivityPoint `json:"points"`
}

type SystemInsights struct {
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
	ErrorRate      float64 `json:"error_rate"`
	Uptime         string  `json:"uptime"`
	MostActiveRole string  `json:"most_active_role"`
}
