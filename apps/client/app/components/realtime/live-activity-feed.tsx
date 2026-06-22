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

export function LiveActivityFeed({
  maxItems = 10,
  className,
}: LiveActivityFeedProps) {
  const activities = useActivityStore((s) => s.activities).slice(0, maxItems);

  if (activities.length === 0) {
    return (
      <div
        className={cn(
          "text-muted-foreground flex flex-col items-center justify-center py-8",
          className,
        )}
      >
        <Activity className="mb-2 h-8 w-8 opacity-40" />
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
            <div
              className={cn(
                "mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full",
                typeColors[a.type],
              )}
            />
            {i < activities.length - 1 && (
              <div className="bg-border w-px flex-1" />
            )}
          </div>
          {/* Content */}
          <div className="min-w-0 flex-1 pb-1">
            <p className="text-body">
              <span className="text-foreground font-semibold">{a.user}</span>{" "}
              <span className="text-muted-foreground">{a.action}</span>{" "}
              {a.target && (
                <span className="text-foreground font-medium">{a.target}</span>
              )}
            </p>
            <p className="text-caption text-muted-foreground">{a.time}</p>
          </div>
        </div>
      ))}
    </div>
  );
}
