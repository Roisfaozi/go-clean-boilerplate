"use client";

import { useEffect, useState } from "react";
import { KPICard } from "~/components/dashboard/kpi-card";
import { Button } from "~/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Icon } from "~/components/shared/icon";
import { usersApi } from "~/lib/api/users";
import { rolesApi } from "~/lib/api/roles";
import { auditApi, AuditLog } from "~/lib/api/audit";
import { statsApi, SystemInsights as SystemInsightsType } from "~/lib/api/stats";
import Link from "next/link";
import { toast } from "sonner";
import { Badge } from "~/components/ui/badge";
import { ActivityChart } from "~/components/dashboard/activity-chart";

export default function DashboardPage() {
  const [stats, setStats] = useState({
    users: 0,
    roles: 0,
    auditLogs: 0,
  });
  const [insights, setInsights] = useState<SystemInsightsType | null>(null);
  const [recentLogs, setRecentLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        // Parallel data fetching
        const [summaryResp, insightsResp, recentLogsResp] =
          await Promise.all([
            statsApi.getSummary(),
            statsApi.getInsights(),
            auditApi.search({
              page: 1,
              page_size: 5,
              sort: [{ colId: "created_at", sort: "desc" }],
            }),
          ]);

        if (summaryResp) {
          setStats({
            users: summaryResp.total_users,
            roles: summaryResp.total_roles,
            auditLogs: summaryResp.total_audit_logs,
          });
        }

        if (insightsResp) {
          setInsights(insightsResp);
        }

        if (recentLogsResp.data) {
          setRecentLogs(recentLogsResp.data);
        }
      } catch (error) {
        console.error("Dashboard fetch error:", error);
        toast.error("Failed to load dashboard data");
      } finally {
        setIsLoading(false);
      }
    };

    fetchDashboardData();
  }, []);

  const formatTimeAgo = (timestamp: number) => {
    const diff = Date.now() - timestamp;
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d ago`;
    if (hours > 0) return `${hours}h ago`;
    if (minutes > 0) return `${minutes}m ago`;
    return "Just now";
  };

  return (
    <div className="space-y-[var(--spacing-gap)]">
      {/* Zone A: KPI Cards */}
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

      <div className="grid grid-cols-1 gap-[var(--spacing-gap)] lg:grid-cols-2">
        <ActivityChart />
        <div className="bg-card text-card-foreground rounded-[var(--radius-lg)] border p-6 shadow-sm">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-primary text-lg font-semibold tracking-tight">
              System Insights
            </h2>
            <Badge variant="outline" className="bg-primary/5">
              Experimental
            </Badge>
          </div>
          <div className="space-y-4">
            <div className="bg-muted/30 rounded-lg border border-dashed p-4">
              <p className="text-muted-foreground text-sm leading-relaxed italic">
                {insights ? (
                  `User engagement is stable. Most active role currently: ${insights.most_active_role}.`
                ) : (
                  "Analyzing system performance and user engagement patterns..."
                )}
              </p>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="rounded-md border p-3">
                <span className="text-muted-foreground text-[10px] font-bold uppercase">
                  Latency
                </span>
                <div className="font-mono text-xl">
                  {isLoading ? "..." : `${insights?.avg_latency_ms || 0}ms`}
                </div>
              </div>
              <div className="rounded-md border p-3">
                <span className="text-muted-foreground text-[10px] font-bold uppercase">
                  Errors
                </span>
                <div className="font-mono text-xl text-emerald-500">
                  {isLoading ? "..." : `${(insights?.error_rate || 0 * 100).toFixed(1)}%`}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Zone B & C: Main Content + Quick Actions */}
      <div className="grid gap-[var(--spacing-gap)] md:grid-cols-7">
        {/* Recent Activity Table (Span 5) */}
        <div className="flex flex-col gap-4 md:col-span-5">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold tracking-tight">
              Recent Activity
            </h2>
            <Link href="/dashboard/audit">
              <Button variant="ghost" size="sm" className="gap-1">
                View All <Icon name="ArrowRight" className="h-4 w-4" />
              </Button>
            </Link>
          </div>

          <div className="bg-card text-card-foreground overflow-hidden rounded-[var(--radius-lg)] border shadow-sm">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>User</TableHead>
                  <TableHead>Action</TableHead>
                  <TableHead>Entity</TableHead>
                  <TableHead>IP</TableHead>
                  <TableHead className="text-right">Time</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      <TableCell>
                        <div className="bg-muted/50 h-4 w-24 animate-pulse rounded" />
                      </TableCell>
                      <TableCell>
                        <div className="bg-muted/50 h-4 w-16 animate-pulse rounded" />
                      </TableCell>
                      <TableCell>
                        <div className="bg-muted/50 h-4 w-20 animate-pulse rounded" />
                      </TableCell>
                      <TableCell>
                        <div className="bg-muted/50 h-4 w-24 animate-pulse rounded" />
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="bg-muted/50 ml-auto h-4 w-12 animate-pulse rounded" />
                      </TableCell>
                    </TableRow>
                  ))
                ) : recentLogs.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={5}
                      className="text-muted-foreground py-8 text-center"
                    >
                      No recent activity found.
                    </TableCell>
                  </TableRow>
                ) : (
                  recentLogs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="text-xs font-medium">
                        {log.user_id}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant="outline"
                          className="bg-muted/50 font-mono text-[10px] uppercase"
                        >
                          {log.action}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-muted-foreground text-xs">
                        {log.entity}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-xs">
                        {log.ip_address}
                      </TableCell>
                      <TableCell className="text-muted-foreground text-right text-xs">
                        {formatTimeAgo(log.created_at)}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>

        {/* Quick Actions (Span 2) */}
        <div className="flex flex-col gap-4 md:col-span-2">
          <h2 className="text-lg font-semibold tracking-tight">
            Quick Actions
          </h2>
          <div className="grid gap-3">
            <Link href="/dashboard/users">
              <Button
                className="h-auto w-full justify-start py-4"
                variant="outline"
              >
                <div className="flex items-center gap-3">
                  <div className="bg-primary/10 text-primary rounded-md p-2">
                    <Icon name="UserPlus" className="h-5 w-5" />
                  </div>
                  <div className="text-left">
                    <div className="font-semibold">Manage Users</div>
                    <div className="text-muted-foreground text-xs">
                      Add or edit accounts
                    </div>
                  </div>
                </div>
              </Button>
            </Link>

            <Link href="/dashboard/roles">
              <Button
                className="h-auto w-full justify-start py-4"
                variant="outline"
              >
                <div className="flex items-center gap-3">
                  <div className="bg-accent/10 text-accent rounded-md p-2">
                    <Icon name="Shield" className="h-5 w-5" />
                  </div>
                  <div className="text-left">
                    <div className="font-semibold">Configure Roles</div>
                    <div className="text-muted-foreground text-xs">
                      Update permissions
                    </div>
                  </div>
                </div>
              </Button>
            </Link>

            <Button
              className="h-auto w-full justify-start py-4"
              variant="outline"
              onClick={() => window.open(auditApi.export(), "_blank")}
            >
              <div className="flex items-center gap-3">
                <div className="bg-secondary/10 text-secondary rounded-md p-2">
                  <Icon name="Download" className="h-5 w-5" />
                </div>
                <div className="text-left">
                  <div className="font-semibold">Export Logs</div>
                  <div className="text-muted-foreground text-xs">
                    Download audit trail
                  </div>
                </div>
              </div>
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
