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
	MostActiveRole string `json:"most_active_role"`
}
