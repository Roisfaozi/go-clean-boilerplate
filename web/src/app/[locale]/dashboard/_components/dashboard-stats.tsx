"use client";

import { useDashboard } from "./dashboard-context";
import { KPICard } from "~/components/dashboard/kpi-card";

export function DashboardStats() {
  const { stats, isLoading } = useDashboard();

  return (
    <div className="grid grid-cols-1 gap-[var(--spacing-gap)] md:grid-cols-2 lg:grid-cols-4">
      <KPICard
        title="Total Users"
        value={isLoading ? "..." : stats.users.toLocaleString()}
        trend={isLoading ? "" : "Active users"}
        trendUp={true}
        iconName="Users"
        description="Registered accounts"
      />
      <KPICard
        title="Defined Roles"
        value={isLoading ? "..." : stats.roles.toLocaleString()}
        trend={isLoading ? "" : "RBAC Policies"}
        trendUp={true}
        iconName="Shield"
        description="Access control roles"
      />
      <KPICard
        title="Total Events"
        value={isLoading ? "..." : stats.auditLogs.toLocaleString()}
        trend={isLoading ? "" : "System logs"}
        trendUp={true}
        iconName="FileText"
        description="Recorded audit trails"
      />
      <KPICard
        title="System Status"
        value="Healthy"
        trend="All systems go"
        trendUp={true}
        iconName="Activity"
        description="No incidents reported"
      />
    </div>
  );
}
