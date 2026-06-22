import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { DashboardGrid } from "@/components/layout/dashboard-grid";
import { NexusButton } from "@casbin/ui";
import { NexusBadge } from "@casbin/ui";
import {
  MetricCard,
  ChartCard,
  AnalyticsCard,
  InsightMetric,
  TrendIndicator,
} from "@/components/data/analytics-cards";
import { Sparkline } from "@/components/charts/charts";
import { useMetricsStore } from "@/stores/realtime-store";
import {
  Users,
  Building2,
  FileText,
  Activity,
  Zap,
  AlertTriangle,
  Clock,
  UserCheck,
  BarChart3,
  TrendingUp,
} from "lucide-react";
import {
  AreaChart,
  Area,
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  PieChart,
  Pie,
  Cell,
} from "recharts";

/* ── Mock data simulating API responses ── */

// /stats/summary
const summaryData = {
  total_users: 2847,
  total_roles: 18,
  total_org_members: 1432,
  total_audit_logs: 54218,
};

const summaryTrends = {
  total_users: { value: 12.5, label: "vs last month" },
  total_roles: { value: 2, label: "new this quarter" },
  total_org_members: { value: 8.2, label: "vs last month" },
  total_audit_logs: { value: -2.1, label: "vs last week" },
};

// /stats/activity
const activityData = [
  { date: "Jan 1", logins: 142, audits: 89 },
  { date: "Jan 2", logins: 165, audits: 102 },
  { date: "Jan 3", logins: 138, audits: 95 },
  { date: "Jan 4", logins: 178, audits: 118 },
  { date: "Jan 5", logins: 192, audits: 125 },
  { date: "Jan 6", logins: 85, audits: 52 },
  { date: "Jan 7", logins: 72, audits: 41 },
  { date: "Jan 8", logins: 155, audits: 98 },
  { date: "Jan 9", logins: 189, audits: 132 },
  { date: "Jan 10", logins: 201, audits: 145 },
  { date: "Jan 11", logins: 175, audits: 112 },
  { date: "Jan 12", logins: 210, audits: 148 },
  { date: "Jan 13", logins: 95, audits: 62 },
  { date: "Jan 14", logins: 88, audits: 55 },
];

// /stats/insights
const insightsData = {
  avg_latency_ms: 42,
  error_rate: 0.12,
  uptime_percent: 99.98,
  most_active_role: "Admin",
};

// Additional derived data
const userGrowthData = [
  { month: "Aug", users: 1820 },
  { month: "Sep", users: 1950 },
  { month: "Oct", users: 2180 },
  { month: "Nov", users: 2410 },
  { month: "Dec", users: 2620 },
  { month: "Jan", users: 2847 },
];

const roleDistribution = [
  { name: "Admin", value: 12, color: "hsl(239, 84%, 67%)" },
  { name: "Editor", value: 35, color: "hsl(168, 76%, 42%)" },
  { name: "Viewer", value: 45, color: "hsl(258, 90%, 66%)" },
  { name: "Auditor", value: 8, color: "hsl(38, 92%, 50%)" },
];

const performanceTimeline = [
  { time: "00:00", latency: 38, errors: 0.08 },
  { time: "04:00", latency: 35, errors: 0.05 },
  { time: "08:00", latency: 52, errors: 0.15 },
  { time: "12:00", latency: 65, errors: 0.22 },
  { time: "16:00", latency: 48, errors: 0.18 },
  { time: "20:00", latency: 42, errors: 0.1 },
  { time: "24:00", latency: 36, errors: 0.06 },
];

const tooltipStyle = {
  background: "hsl(var(--popover))",
  border: "1px solid hsl(var(--border))",
  borderRadius: 8,
  fontSize: 13,
};

type Range = "7d" | "14d" | "30d";

export default function DashboardPage() {
  const [range, setRange] = useState<Range>("14d");
  const { metrics, history } = useMetricsStore();

  const displayedActivity =
    range === "7d" ? activityData.slice(7) : activityData;

  // Real-time RPS history for sparklines
  const rpsHistory = history.map((h) => h.rps);
  const _latencyHistory = history.map((h) => h.latency);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Dashboard"
        description="Overview of workspace activity, performance, and growth metrics."
        actions={
          <div className="flex items-center gap-2">
            {(["7d", "14d", "30d"] as Range[]).map((r) => (
              <NexusButton
                key={r}
                variant={range === r ? "primary" : "outline"}
                size="sm"
                onClick={() => setRange(r)}
              >
                {r}
              </NexusButton>
            ))}
          </div>
        }
      />

      {/* ═══ Overview: StatCards ═══ */}
      <DashboardGrid columns={4}>
        <MetricCard
          title="Total Users"
          value={
            metrics?.total_users.toLocaleString() ||
            summaryData.total_users.toLocaleString()
          }
          trend={summaryTrends.total_users}
          icon={Users}
          iconColor="bg-primary/10"
          sparkline={
            <Sparkline
              data={
                rpsHistory.length > 0
                  ? rpsHistory
                  : [18, 22, 25, 23, 28, 30, 28]
              }
              color="hsl(239, 84%, 67%)"
            />
          }
        />
        <MetricCard
          title="Active Users"
          value={metrics?.active_users.toString() || "0"}
          trend={{ value: 5.2, label: "Live now" }}
          icon={Activity}
          iconColor="bg-success/10"
          sparkline={
            <Sparkline
              data={
                rpsHistory.length > 0
                  ? rpsHistory
                  : [14, 14, 15, 16, 16, 17, 18]
              }
              color="hsl(168, 76%, 42%)"
            />
          }
        />
        <MetricCard
          title="Org Members"
          value={summaryData.total_org_members.toLocaleString()}
          trend={summaryTrends.total_org_members}
          icon={Building2}
          iconColor="bg-secondary/10"
          sparkline={
            <Sparkline
              data={[1100, 1180, 1250, 1300, 1350, 1400, 1432]}
              color="hsl(168, 76%, 42%)"
            />
          }
        />
        <MetricCard
          title="Audit Logs"
          value={summaryData.total_audit_logs.toLocaleString()}
          trend={summaryTrends.total_audit_logs}
          icon={FileText}
          iconColor="bg-warning/10"
          sparkline={
            <Sparkline
              data={[560, 540, 580, 520, 550, 530, 542]}
              color="hsl(38, 92%, 50%)"
            />
          }
        />
      </DashboardGrid>

      {/* ═══ Activity Analytics ═══ */}
      <div className="gap-gap grid grid-cols-1 lg:grid-cols-3">
        <ChartCard
          title="Activity Analytics"
          description="Daily logins and audit events"
          className="lg:col-span-2"
          height={320}
          actions={
            <NexusBadge variant="info">
              <Activity className="mr-1 h-3 w-3" />
              Live
            </NexusBadge>
          }
        >
          <AreaChart data={displayedActivity}>
            <defs>
              <linearGradient id="loginGrad" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="0%"
                  stopColor="hsl(239, 84%, 67%)"
                  stopOpacity={0.25}
                />
                <stop
                  offset="100%"
                  stopColor="hsl(239, 84%, 67%)"
                  stopOpacity={0}
                />
              </linearGradient>
              <linearGradient id="auditGrad" x1="0" y1="0" x2="0" y2="1">
                <stop
                  offset="0%"
                  stopColor="hsl(168, 76%, 42%)"
                  stopOpacity={0.25}
                />
                <stop
                  offset="100%"
                  stopColor="hsl(168, 76%, 42%)"
                  stopOpacity={0}
                />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
            <XAxis
              dataKey="date"
              tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
            />
            <YAxis
              tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
            />
            <Tooltip contentStyle={tooltipStyle} />
            <Legend />
            <Area
              type="monotone"
              dataKey="logins"
              name="Daily Logins"
              stroke="hsl(239, 84%, 67%)"
              fill="url(#loginGrad)"
              strokeWidth={2}
            />
            <Area
              type="monotone"
              dataKey="audits"
              name="Audit Events"
              stroke="hsl(168, 76%, 42%)"
              fill="url(#auditGrad)"
              strokeWidth={2}
            />
          </AreaChart>
        </ChartCard>

        {/* System Insights (sidebar) */}
        <AnalyticsCard
          title="System Performance"
          description="Real-time health metrics"
          actions={
            <NexusBadge variant="success" className="animate-pulse">
              LIVE
            </NexusBadge>
          }
        >
          <div className="space-y-0">
            <InsightMetric
              label="Real-time RPS"
              value={`${metrics?.rps.toFixed(1) || "0.0"}`}
              status={(metrics?.rps || 0) < 50 ? "good" : "warning"}
              description="Requests per second"
            />
            <InsightMetric
              label="Avg Latency"
              value={`${metrics?.avg_latency.toFixed(0) || insightsData.avg_latency_ms}ms`}
              status={(metrics?.avg_latency || 0) < 100 ? "good" : "warning"}
              description="Response time p95"
            />
            <InsightMetric
              label="Error Rate"
              value={`${metrics?.error_rate || insightsData.error_rate}%`}
              status={(metrics?.error_rate || 0) < 0.5 ? "good" : "critical"}
              description="Last 24 hours"
            />
            <InsightMetric
              label="Uptime"
              value={`${metrics?.uptime || insightsData.uptime_percent}%`}
              status={(metrics?.uptime || 0) > 99.9 ? "good" : "warning"}
              description="30-day rolling"
            />
          </div>

          {/* Mini performance gauges */}
          <div className="mt-4 grid grid-cols-3 gap-3">
            <div className="space-y-1 text-center">
              <div className="bg-success/10 inline-flex h-10 w-10 items-center justify-center rounded-full">
                <Zap className="text-success h-4 w-4" />
              </div>
              <p className="text-caption text-foreground font-medium">
                {insightsData.avg_latency_ms}ms
              </p>
              <p className="text-caption text-muted-foreground">Latency</p>
            </div>
            <div className="space-y-1 text-center">
              <div className="bg-danger/10 inline-flex h-10 w-10 items-center justify-center rounded-full">
                <AlertTriangle className="text-danger h-4 w-4" />
              </div>
              <p className="text-caption text-foreground font-medium">
                {insightsData.error_rate}%
              </p>
              <p className="text-caption text-muted-foreground">Errors</p>
            </div>
            <div className="space-y-1 text-center">
              <div className="bg-primary/10 inline-flex h-10 w-10 items-center justify-center rounded-full">
                <Clock className="text-primary h-4 w-4" />
              </div>
              <p className="text-caption text-foreground font-medium">
                {insightsData.uptime_percent}%
              </p>
              <p className="text-caption text-muted-foreground">Uptime</p>
            </div>
          </div>
        </AnalyticsCard>
      </div>

      {/* ═══ User Insights & Performance ═══ */}
      <div className="gap-gap grid grid-cols-1 lg:grid-cols-2">
        {/* User Growth */}
        <ChartCard
          title="User Growth"
          description="Monthly registered users"
          height={280}
        >
          <BarChart data={userGrowthData}>
            <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
            <XAxis
              dataKey="month"
              tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
            />
            <YAxis
              tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
            />
            <Tooltip contentStyle={tooltipStyle} />
            <Bar
              dataKey="users"
              name="Users"
              fill="hsl(239, 84%, 67%)"
              radius={[6, 6, 0, 0]}
            />
          </BarChart>
        </ChartCard>

        {/* Performance Timeline */}
        <ChartCard
          title="Performance Timeline"
          description="Latency & error rate (24h)"
          height={280}
        >
          <LineChart data={performanceTimeline}>
            <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
            <XAxis
              dataKey="time"
              tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
            />
            <YAxis
              yAxisId="left"
              tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
            />
            <YAxis
              yAxisId="right"
              orientation="right"
              tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
            />
            <Tooltip contentStyle={tooltipStyle} />
            <Legend />
            <Line
              yAxisId="left"
              type="monotone"
              dataKey="latency"
              name="Latency (ms)"
              stroke="hsl(239, 84%, 67%)"
              strokeWidth={2}
              dot={false}
            />
            <Line
              yAxisId="right"
              type="monotone"
              dataKey="errors"
              name="Error Rate (%)"
              stroke="hsl(0, 72%, 51%)"
              strokeWidth={2}
              dot={false}
              strokeDasharray="5 5"
            />
          </LineChart>
        </ChartCard>
      </div>

      {/* ═══ Bottom Row: Role Distribution + Quick Metrics ═══ */}
      <div className="gap-gap grid grid-cols-1 lg:grid-cols-3">
        <ChartCard
          title="Role Distribution"
          description="Active user roles"
          height={260}
        >
          <PieChart>
            <Pie
              data={roleDistribution}
              cx="50%"
              cy="50%"
              innerRadius="50%"
              outerRadius="80%"
              paddingAngle={3}
              dataKey="value"
            >
              {roleDistribution.map((entry, i) => (
                <Cell key={i} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip contentStyle={tooltipStyle} />
            <Legend />
          </PieChart>
        </ChartCard>

        <AnalyticsCard
          title="Growth Metrics"
          description="Key performance indicators"
          className="lg:col-span-2"
        >
          <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
            {[
              {
                label: "New Users",
                value: "347",
                trend: 15.2,
                icon: UserCheck,
              },
              {
                label: "Active Sessions",
                value: "1,204",
                trend: 8.7,
                icon: Activity,
              },
              {
                label: "API Calls",
                value: "284K",
                trend: -3.4,
                icon: BarChart3,
              },
              {
                label: "Conversions",
                value: "12.4%",
                trend: 22.1,
                icon: TrendingUp,
              },
            ].map((metric) => (
              <div
                key={metric.label}
                className="bg-surface space-y-2 rounded-lg p-4"
              >
                <div className="flex items-center justify-between">
                  <metric.icon className="text-muted-foreground h-4 w-4" />
                  <TrendIndicator value={metric.trend} size="sm" />
                </div>
                <p className="text-h2 text-foreground font-bold">
                  {metric.value}
                </p>
                <p className="text-caption text-muted-foreground">
                  {metric.label}
                </p>
              </div>
            ))}
          </div>

          {/* Recent activity summary */}
          <div className="border-border mt-4 border-t pt-4">
            <div className="flex items-center justify-between">
              <p className="text-small text-muted-foreground">
                Platform health score
              </p>
              <div className="flex items-center gap-2">
                <div className="bg-muted h-2 w-32 overflow-hidden rounded-full">
                  <div
                    className="bg-success h-full rounded-full"
                    style={{ width: "94%" }}
                  />
                </div>
                <span className="text-body text-success font-semibold">
                  94%
                </span>
              </div>
            </div>
          </div>
        </AnalyticsCard>
      </div>
    </div>
  );
}
