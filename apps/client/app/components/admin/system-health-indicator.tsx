import { cn } from "@casbin/ui";
import { NexusCard } from "@casbin/ui";
import { Badge } from "@casbin/ui";
import {
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Activity,
  Database,
  Globe,
  Cpu,
  HardDrive,
  Clock,
} from "lucide-react";

export interface ServiceHealth {
  name: string;
  status: "operational" | "degraded" | "down";
  latency_ms: number;
  uptime: number;
  last_check: string;
  icon?: React.ElementType;
}

const statusConfig = {
  operational: {
    label: "Operational",
    icon: CheckCircle2,
    color: "text-success",
    bg: "bg-success/10",
    badge: "secondary" as const,
    dot: "bg-success",
  },
  degraded: {
    label: "Degraded",
    icon: AlertTriangle,
    color: "text-warning",
    bg: "bg-warning/10",
    badge: "outline" as const,
    dot: "bg-warning",
  },
  down: {
    label: "Down",
    icon: XCircle,
    color: "text-destructive",
    bg: "bg-destructive/10",
    badge: "destructive" as const,
    dot: "bg-destructive",
  },
};

function HealthBar({ value }: { value: number }) {
  const color =
    value >= 99.9
      ? "bg-success"
      : value >= 99
        ? "bg-warning"
        : "bg-destructive";
  return (
    <div className="bg-muted h-1.5 w-full overflow-hidden rounded-full">
      <div
        className={cn("h-full rounded-full transition-all", color)}
        style={{ width: `${value}%` }}
      />
    </div>
  );
}

interface SystemHealthIndicatorProps {
  services: ServiceHealth[];
  overallStatus?: "operational" | "degraded" | "down";
}

export function SystemHealthIndicator({
  services,
  overallStatus,
}: SystemHealthIndicatorProps) {
  const overall =
    overallStatus ??
    (services.some((s) => s.status === "down")
      ? "down"
      : services.some((s) => s.status === "degraded")
        ? "degraded"
        : "operational");

  const cfg = statusConfig[overall];
  const OverallIcon = cfg.icon;

  const defaultIcons: Record<string, React.ElementType> = {
    "API Gateway": Globe,
    Database: Database,
    "Auth Service": Cpu,
    Storage: HardDrive,
    "Background Jobs": Activity,
  };

  return (
    <div className="space-y-4">
      {/* Overall status banner */}
      <NexusCard className={cn("flex items-center gap-4", cfg.bg, "border-0")}>
        <div
          className={cn(
            "flex h-12 w-12 items-center justify-center rounded-xl",
            cfg.bg,
          )}
        >
          <OverallIcon className={cn("h-6 w-6", cfg.color)} />
        </div>
        <div className="flex-1">
          <p className="text-muted-foreground text-sm font-medium">
            System Status
          </p>
          <p className={cn("text-lg font-bold", cfg.color)}>{cfg.label}</p>
        </div>
        <Badge variant={cfg.badge} className="gap-1.5">
          <span className={cn("h-2 w-2 animate-pulse rounded-full", cfg.dot)} />
          {cfg.label}
        </Badge>
      </NexusCard>

      {/* Service list */}
      <div className="space-y-2">
        {services.map((service) => {
          const sc = statusConfig[service.status];
          const ServiceIcon =
            service.icon ?? defaultIcons[service.name] ?? Activity;
          return (
            <NexusCard
              key={service.name}
              className="flex items-center gap-4 py-3"
            >
              <div
                className={cn(
                  "flex h-9 w-9 shrink-0 items-center justify-center rounded-lg",
                  sc.bg,
                )}
              >
                <ServiceIcon className={cn("h-4 w-4", sc.color)} />
              </div>
              <div className="min-w-0 flex-1">
                <div className="mb-1 flex items-center justify-between">
                  <span className="text-foreground text-sm font-medium">
                    {service.name}
                  </span>
                  <Badge
                    variant={sc.badge}
                    className="gap-1 px-1.5 py-0 text-[10px]"
                  >
                    <span className={cn("h-1.5 w-1.5 rounded-full", sc.dot)} />
                    {sc.label}
                  </Badge>
                </div>
                <HealthBar value={service.uptime} />
                <div className="text-muted-foreground mt-1.5 flex items-center gap-4 text-xs">
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    {service.latency_ms}ms
                  </span>
                  <span>{service.uptime}% uptime</span>
                  <span className="ml-auto">Checked {service.last_check}</span>
                </div>
              </div>
            </NexusCard>
          );
        })}
      </div>
    </div>
  );
}
