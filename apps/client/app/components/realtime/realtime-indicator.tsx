import { cn } from "@casbin/ui";
import { useConnectionStore } from "@/stores/realtime-store";
import { Tooltip, TooltipContent, TooltipTrigger } from "@casbin/ui";

interface RealtimeIndicatorProps {
  showLabel?: boolean;
  className?: string;
}

export function RealtimeIndicator({
  showLabel,
  className,
}: RealtimeIndicatorProps) {
  const { sseConnected, wsConnected } = useConnectionStore();
  const allConnected = sseConnected && wsConnected;
  const partialConnected = sseConnected || wsConnected;

  const status = allConnected
    ? "connected"
    : partialConnected
      ? "partial"
      : "disconnected";
  const color =
    status === "connected"
      ? "text-success"
      : status === "partial"
        ? "text-warning"
        : "text-danger";
  const dotColor =
    status === "connected"
      ? "bg-success"
      : status === "partial"
        ? "bg-warning"
        : "bg-danger";
  const label =
    status === "connected"
      ? "Live"
      : status === "partial"
        ? "Partial"
        : "Offline";
  const details = `SSE: ${sseConnected ? "✓" : "✗"} | WS: ${wsConnected ? "✓" : "✗"}`;

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span
          className={cn(
            "inline-flex cursor-default items-center gap-1.5",
            className,
          )}
        >
          <span className="relative flex h-2 w-2">
            {(allConnected || partialConnected) && (
              <span
                className={cn(
                  "absolute inset-0 animate-ping rounded-full opacity-75",
                  dotColor,
                )}
              />
            )}
            <span
              className={cn(
                "relative inline-flex h-2 w-2 rounded-full",
                dotColor,
              )}
            />
          </span>
          {showLabel && (
            <span className={cn("text-caption font-medium", color)}>
              {label}
            </span>
          )}
        </span>
      </TooltipTrigger>
      <TooltipContent>{details}</TooltipContent>
    </Tooltip>
  );
}
