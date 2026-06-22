import { cn } from "./lib/utils";
import { cva, type VariantProps } from "class-variance-authority";
import {
  AlertCircle,
  CheckCircle2,
  Info,
  AlertTriangle,
  X,
} from "lucide-react";
import { useState } from "react";

/* ── InlineAlert ── */
const alertVariants = cva(
  "flex items-start gap-3 rounded-lg border p-4 text-body",
  {
    variants: {
      variant: {
        info: "bg-info/5 border-info/20 text-info",
        success: "bg-success/5 border-success/20 text-success",
        warning: "bg-warning/5 border-warning/20 text-warning",
        danger: "bg-danger/5 border-danger/20 text-danger",
      },
    },
    defaultVariants: { variant: "info" },
  },
);

const iconMap = {
  info: Info,
  success: CheckCircle2,
  warning: AlertTriangle,
  danger: AlertCircle,
};

interface InlineAlertProps extends VariantProps<typeof alertVariants> {
  title?: string;
  children: React.ReactNode;
  dismissible?: boolean;
  className?: string;
}

export function InlineAlert({
  variant = "info",
  title,
  children,
  dismissible,
  className,
}: InlineAlertProps) {
  const [visible, setVisible] = useState(true);
  if (!visible) return null;
  const Icon = iconMap[variant || "info"];

  return (
    <div className={cn(alertVariants({ variant }), className)}>
      <Icon className="mt-0.5 h-5 w-5 shrink-0" />
      <div className="flex-1 space-y-1">
        {title && <p className="font-semibold">{title}</p>}
        <div className="text-foreground">{children}</div>
      </div>
      {dismissible && (
        <button
          onClick={() => setVisible(false)}
          className="shrink-0 hover:opacity-70"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}

/* ── StatusIndicator ── */
const statusColors = {
  online: "bg-success",
  offline: "bg-muted-foreground",
  busy: "bg-danger",
  away: "bg-warning",
};

interface StatusIndicatorProps {
  status: keyof typeof statusColors;
  label?: string;
  pulse?: boolean;
  className?: string;
}

export function StatusIndicator({
  status,
  label,
  pulse = true,
  className,
}: StatusIndicatorProps) {
  return (
    <span className={cn("inline-flex items-center gap-2", className)}>
      <span className="relative flex h-2.5 w-2.5">
        {pulse && status === "online" && (
          <span
            className={cn(
              "absolute inset-0 animate-ping rounded-full opacity-75",
              statusColors[status],
            )}
          />
        )}
        <span
          className={cn(
            "relative inline-flex h-2.5 w-2.5 rounded-full",
            statusColors[status],
          )}
        />
      </span>
      {label && (
        <span className="text-small text-foreground capitalize">
          {label || status}
        </span>
      )}
    </span>
  );
}

/* ── ProgressBar ── */
const progressBarVariants = cva(
  "h-full rounded-full transition-all duration-slow",
  {
    variants: {
      variant: {
        primary: "bg-primary",
        success: "bg-success",
        warning: "bg-warning",
        danger: "bg-danger",
        info: "bg-info",
      },
    },
    defaultVariants: { variant: "primary" },
  },
);

interface ProgressBarProps extends VariantProps<typeof progressBarVariants> {
  value: number;
  max?: number;
  label?: string;
  showValue?: boolean;
  size?: "sm" | "md" | "lg";
  className?: string;
}

export function ProgressBar({
  value,
  max = 100,
  label,
  showValue,
  variant,
  size = "md",
  className,
}: ProgressBarProps) {
  const pct = Math.min(100, Math.max(0, (value / max) * 100));
  const h = size === "sm" ? "h-1.5" : size === "lg" ? "h-4" : "h-2.5";

  return (
    <div className={cn("space-y-1.5", className)}>
      {(label || showValue) && (
        <div className="text-small flex items-center justify-between">
          {label && (
            <span className="text-foreground font-medium">{label}</span>
          )}
          {showValue && (
            <span className="text-muted-foreground">{Math.round(pct)}%</span>
          )}
        </div>
      )}
      <div className={cn("bg-muted w-full overflow-hidden rounded-full", h)}>
        <div
          className={cn(progressBarVariants({ variant }), h)}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}

/* ── SkeletonLoader ── */
interface SkeletonLoaderProps {
  variant?: "text" | "circle" | "card" | "table";
  rows?: number;
  className?: string;
}

export function SkeletonLoader({
  variant = "text",
  rows = 3,
  className,
}: SkeletonLoaderProps) {
  const base = "animate-pulse rounded-md bg-muted";

  if (variant === "circle") {
    return <div className={cn(base, "h-10 w-10 rounded-full", className)} />;
  }

  if (variant === "card") {
    return (
      <div
        className={cn(
          "border-border p-card-pad space-y-3 rounded-lg border",
          className,
        )}
      >
        <div className={cn(base, "h-4 w-1/3")} />
        <div className={cn(base, "h-8 w-2/3")} />
        <div className={cn(base, "h-3 w-full")} />
        <div className={cn(base, "h-3 w-4/5")} />
      </div>
    );
  }

  if (variant === "table") {
    return (
      <div className={cn("space-y-2", className)}>
        <div className={cn(base, "h-10 w-full")} />
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className={cn(base, "h-12 w-full")} />
        ))}
      </div>
    );
  }

  // text
  return (
    <div className={cn("space-y-2", className)}>
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className={cn(base, "h-4", i === rows - 1 ? "w-3/4" : "w-full")}
        />
      ))}
    </div>
  );
}

/* ── NotificationCenter ── */
export interface Notification {
  id: string;
  title: string;
  description?: string;
  time: string;
  read?: boolean;
  icon?: React.ReactNode;
}

interface NotificationCenterProps {
  notifications: Notification[];
  onRead?: (id: string) => void;
  onClear?: () => void;
  className?: string;
}

export function NotificationCenter({
  notifications,
  onRead,
  onClear,
  className,
}: NotificationCenterProps) {
  const unread = notifications.filter((n) => !n.read).length;

  return (
    <div
      className={cn(
        "bg-popover border-border w-full max-w-sm overflow-hidden rounded-lg border shadow-lg",
        className,
      )}
    >
      <div className="border-border flex items-center justify-between border-b px-4 py-3">
        <div className="flex items-center gap-2">
          <h3 className="text-h4 text-foreground">Notifications</h3>
          {unread > 0 && (
            <span className="bg-danger text-danger-foreground text-caption inline-flex h-5 min-w-5 items-center justify-center rounded-full px-1.5">
              {unread}
            </span>
          )}
        </div>
        {onClear && (
          <button
            onClick={onClear}
            className="text-caption text-primary hover:underline"
          >
            Clear all
          </button>
        )}
      </div>
      <div className="divide-border max-h-80 divide-y overflow-y-auto">
        {notifications.length === 0 ? (
          <p className="text-body text-muted-foreground p-6 text-center">
            No notifications
          </p>
        ) : (
          notifications.map((n) => (
            <button
              key={n.id}
              onClick={() => onRead?.(n.id)}
              className={cn(
                "hover:bg-surface-hover flex w-full items-start gap-3 px-4 py-3 text-left transition-colors",
                !n.read && "bg-primary/5",
              )}
            >
              {n.icon && <span className="mt-0.5 shrink-0">{n.icon}</span>}
              <div className="min-w-0 flex-1">
                <p
                  className={cn(
                    "text-body truncate",
                    !n.read
                      ? "text-foreground font-semibold"
                      : "text-foreground",
                  )}
                >
                  {n.title}
                </p>
                {n.description && (
                  <p className="text-caption text-muted-foreground truncate">
                    {n.description}
                  </p>
                )}
              </div>
              <span className="text-caption text-muted-foreground shrink-0">
                {n.time}
              </span>
            </button>
          ))
        )}
      </div>
    </div>
  );
}
