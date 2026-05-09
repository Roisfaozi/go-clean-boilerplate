import { PageHeader } from "@/components/layout/page-header";
import { SystemHealthIndicator, ServiceHealth } from "@/components/admin/system-health-indicator";
import { MetricsPanel, MetricSummary, TimeSeriesPoint } from "@/components/admin/metrics-panel";
import { Cpu, HardDrive, Wifi, Clock } from "lucide-react";

const services: ServiceHealth[] = [
  { name: "API Gateway", status: "operational", latency_ms: 32, uptime: 99.99, last_check: "12s ago" },
  { name: "Database", status: "operational", latency_ms: 8, uptime: 99.98, last_check: "12s ago" },
  { name: "Auth Service", status: "degraded", latency_ms: 145, uptime: 99.82, last_check: "30s ago" },
  { name: "Storage", status: "operational", latency_ms: 52, uptime: 99.95, last_check: "1m ago" },
  { name: "Background Jobs", status: "operational", latency_ms: 15, uptime: 99.97, last_check: "45s ago" },
];

const resourceSummaries: MetricSummary[] = [
  { label: "CPU Usage", value: "34%", change: -5.2, changeLabel: "vs last hour", icon: Cpu },
  { label: "Memory", value: "6.2 GB", change: 2.1, changeLabel: "vs last hour", icon: HardDrive },
  { label: "Network I/O", value: "142 MB/s", change: 8.4, changeLabel: "vs last hour", icon: Wifi },
  { label: "Avg Response", value: "42ms", change: -12.3, changeLabel: "vs last hour", icon: Clock },
];

const resourceTimeline: TimeSeriesPoint[] = [
  { label: "00:00", value: 28, value2: 35 },
  { label: "04:00", value: 22, value2: 30 },
  { label: "08:00", value: 45, value2: 52 },
  { label: "12:00", value: 62, value2: 68 },
  { label: "16:00", value: 48, value2: 55 },
  { label: "20:00", value: 35, value2: 42 },
  { label: "Now", value: 34, value2: 38 },
];

export default function SystemHealthPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="System Health"
        description="Real-time service status and infrastructure monitoring"
      />
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <SystemHealthIndicator services={services} />
        <MetricsPanel
          title="Resource Usage"
          description="CPU and memory over the last 24 hours"
          summaries={resourceSummaries}
          chartData={resourceTimeline}
          chartType="line"
          chartLabel="CPU %"
          chartLabel2="Memory %"
        />
      </div>
    </div>
  );
}
