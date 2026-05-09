import { cn } from "@casbin/ui";
import { usePresenceStore, type PresenceUser } from "@/stores/realtime-store";
import {  Tooltip, TooltipContent, TooltipTrigger  } from "@casbin/ui";

const statusColors: Record<PresenceUser["status"], string> = {
  online: "bg-success",
  away: "bg-warning",
  busy: "bg-danger",
};

interface PresenceAvatarsProps {
  max?: number;
  size?: "sm" | "md" | "lg";
  showCount?: boolean;
  className?: string;
}

const sizeMap = {
  sm: "h-7 w-7 text-caption",
  md: "h-9 w-9 text-small",
  lg: "h-11 w-11 text-body",
};

const dotSize = {
  sm: "h-2 w-2",
  md: "h-2.5 w-2.5",
  lg: "h-3 w-3",
};

export function PresenceAvatars({ max = 5, size = "md", showCount = true, className }: PresenceAvatarsProps) {
  const users = usePresenceStore((s) => s.users);
  const onlineCount = usePresenceStore((s) => s.onlineCount);
  const displayed = users.slice(0, max);
  const overflow = users.length - max;

  return (
    <div className={cn("flex items-center", className)}>
      <div className="flex -space-x-2">
        {displayed.map((user) => (
          <Tooltip key={user.id}>
            <TooltipTrigger asChild>
              <div className="relative">
                {user.avatar ? (
                  <img
                    src={user.avatar}
                    alt={user.name}
                    className={cn("rounded-full border-2 border-background object-cover", sizeMap[size])}
                  />
                ) : (
                  <div
                    className={cn(
                      "rounded-full border-2 border-background bg-primary/10 text-primary font-semibold flex items-center justify-center",
                      sizeMap[size]
                    )}
                  >
                    {user.name.charAt(0).toUpperCase()}
                  </div>
                )}
                <span
                  className={cn(
                    "absolute bottom-0 right-0 rounded-full border-2 border-background",
                    dotSize[size],
                    statusColors[user.status]
                  )}
                />
              </div>
            </TooltipTrigger>
            <TooltipContent>
              {user.name} — {user.status}
            </TooltipContent>
          </Tooltip>
        ))}
        {overflow > 0 && (
          <div
            className={cn(
              "rounded-full border-2 border-background bg-muted text-muted-foreground font-semibold flex items-center justify-center",
              sizeMap[size]
            )}
          >
            +{overflow}
          </div>
        )}
      </div>
      {showCount && (
        <span className="ml-3 text-small text-muted-foreground">
          <span className="font-semibold text-foreground">{onlineCount}</span> online
        </span>
      )}
    </div>
  );
}
