import { PageHeader } from "@/components/layout/page-header";
import {
  MetricsPanel,
  type MetricSummary,
  type TimeSeriesPoint,
} from "@/components/admin/metrics-panel";
import { NexusCard, NexusCardHeader, NexusCardTitle } from "@casbin/ui";
import { Badge } from "@casbin/ui";
import {
  Users,
  Shield,
  FileText,
  Zap,
  TrendingUp,
  Activity,
} from "lucide-react";

const requestMetrics: MetricSummary[] = [
  {
    label: "Total Requests",
    value: "1.24M",
    change: 14.2,
    changeLabel: "vs last week",
    icon: Activity,
  },
  {
    label: "Avg Latency",
    value: "42ms",
    change: -8.1,
    changeLabel: "vs last week",
    icon: Zap,
  },
  {
    label: "Error Rate",
    value: "0.12%",
    change: -22.5,
    changeLabel: "vs last week",
    icon: TrendingUp,
  },
  {
    label: "Active Users",
    value: "1,847",
    change: 6.3,
    changeLabel: "vs last week",
    icon: Users,
  },
];

const requestTimeline: TimeSeriesPoint[] = [
  { label: "Mon", value: 184000 },
  { label: "Tue", value: 195000 },
  { label: "Wed", value: 210000 },
  { label: "Thu", value: 178000 },
  { label: "Fri", value: 202000 },
  { label: "Sat", value: 132000 },
  { label: "Sun", value: 118000 },
];

const securityMetrics: MetricSummary[] = [
  {
    label: "Failed Logins",
    value: "47",
    change: -15.3,
    changeLabel: "vs last week",
    icon: Shield,
  },
  {
    label: "Role Changes",
    value: "12",
    change: 33.0,
    changeLabel: "vs last week",
    icon: Shield,
  },
  {
    label: "New Permissions",
    value: "8",
    change: 0,
    changeLabel: "this week",
    icon: Shield,
  },
  {
    label: "Audit Events",
    value: "3,421",
    change: 5.8,
    changeLabel: "vs last week",
    icon: FileText,
  },
];

const securityTimeline: TimeSeriesPoint[] = [
  { label: "Mon", value: 520, value2: 8 },
  { label: "Tue", value: 480, value2: 12 },
  { label: "Wed", value: 540, value2: 5 },
  { label: "Thu", value: 490, value2: 9 },
  { label: "Fri", value: 510, value2: 7 },
  { label: "Sat", value: 380, value2: 3 },
  { label: "Sun", value: 350, value2: 3 },
];

const topEndpoints = [
  {
    endpoint: "GET /api/users",
    calls: "42.3K",
    p99: "38ms",
    status: "healthy",
  },
  {
    endpoint: "POST /api/auth/login",
    calls: "28.1K",
    p99: "125ms",
    status: "healthy",
  },
  {
    endpoint: "GET /api/projects",
    calls: "18.7K",
    p99: "45ms",
    status: "healthy",
  },
  {
    endpoint: "PUT /api/roles/:id",
    calls: "3.2K",
    p99: "210ms",
    status: "warning",
  },
  {
    endpoint: "DELETE /api/permissions",
    calls: "1.1K",
    p99: "92ms",
    status: "healthy",
  },
];

export default function SystemInsightsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="System Insights"
        description="Analytics, performance breakdowns, and security overview"
      />

      <MetricsPanel
        title="Request Volume"
        description="API request volume and latency trends"
        summaries={requestMetrics}
        chartData={requestTimeline}
        chartType="bar"
        chartLabel="Requests"
      />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <MetricsPanel
          title="Security Events"
          description="Audit events and failed login attempts"
          summaries={securityMetrics}
          chartData={securityTimeline}
          chartType="area"
          chartLabel="Audit Events"
          chartLabel2="Failed Logins"
        />

        <NexusCard>
          <NexusCardHeader>
            <NexusCardTitle>Top Endpoints</NexusCardTitle>
          </NexusCardHeader>
          <div className="space-y-2">
            {topEndpoints.map((ep) => (
              <div
                key={ep.endpoint}
                className="bg-muted/50 hover:bg-muted/80 flex items-center justify-between rounded-lg p-3 transition-colors"
              >
                <div className="min-w-0">
                  <p className="text-foreground truncate font-mono text-sm">
                    {ep.endpoint}
                  </p>
                  <p className="text-muted-foreground text-xs">
                    {ep.calls} calls · p99: {ep.p99}
                  </p>
                </div>
                <Badge
                  variant={ep.status === "healthy" ? "secondary" : "outline"}
                >
                  {ep.status}
                </Badge>
              </div>
            ))}
          </div>
        </NexusCard>
      </div>
    </div>
  );
}
