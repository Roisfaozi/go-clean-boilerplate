import { cn } from "@casbin/ui";
import { useActivityStore, type LiveActivity } from "@/stores/realtime-store";
import { Activity } from "lucide-react";

const typeColors: Record<LiveActivity["type"], string> = {
  info: "bg-info",
  success: "bg-success",
  warning: "bg-warning",
  danger: "bg-danger",
};

interface LiveActivityFeedProps {
  maxItems?: number;
  className?: string;
}

export function LiveActivityFeed({ maxItems = 10, className }: LiveActivityFeedProps) {
  const activities = useActivityStore((s) => s.activities).slice(0, maxItems);

  if (activities.length === 0) {
    return (
      <div className={cn("flex flex-col items-center justify-center py-8 text-muted-foreground", className)}>
        <Activity className="h-8 w-8 mb-2 opacity-40" />
        <p className="text-body">No recent activity</p>
      </div>
    );
  }

  return (
    <div className={cn("space-y-0", className)}>
      {activities.map((a, i) => (
        <div key={a.id} className="flex gap-3 py-2.5">
          {/* Timeline dot + line */}
          <div className="flex flex-col items-center">
            <div className={cn("h-2.5 w-2.5 rounded-full shrink-0 mt-1.5", typeColors[a.type])} />
            {i < activities.length - 1 && <div className="w-px flex-1 bg-border" />}
          </div>
          {/* Content */}
          <div className="flex-1 min-w-0 pb-1">
            <p className="text-body">
              <span className="font-semibold text-foreground">{a.user}</span>{" "}
              <span className="text-muted-foreground">{a.action}</span>{" "}
              {a.target && <span className="font-medium text-foreground">{a.target}</span>}
            </p>
            <p className="text-caption text-muted-foreground">{a.time}</p>
          </div>
        </div>
      ))}
    </div>
  );
}
