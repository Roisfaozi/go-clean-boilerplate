import { useState, useRef, useEffect } from "react";
import { cn } from "@casbin/ui";
import {
  useNotificationStore,
  type RealtimeNotification,
} from "@/stores/realtime-store";
import {
  Bell,
  CheckCheck,
  Trash2,
  Info,
  CheckCircle2,
  AlertTriangle,
  AlertCircle,
  Radio,
} from "lucide-react";
import { NexusButton } from "@casbin/ui";

const typeIcons: Record<RealtimeNotification["type"], React.ReactNode> = {
  info: <Info className="text-info h-4 w-4" />,
  success: <CheckCircle2 className="text-success h-4 w-4" />,
  warning: <AlertTriangle className="text-warning h-4 w-4" />,
  danger: <AlertCircle className="text-danger h-4 w-4" />,
  system: <Radio className="text-primary h-4 w-4" />,
};

interface NotificationBellProps {
  className?: string;
}

export function NotificationBell({ className }: NotificationBellProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const { notifications, unreadCount, markAsRead, markAllAsRead, clearAll } =
    useNotificationStore();

  // Close on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node))
        setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  return (
    <div ref={ref} className={cn("relative", className)}>
      <button
        onClick={() => setOpen(!open)}
        className="hover:bg-surface-hover text-muted-foreground relative rounded-md p-2 transition-colors"
      >
        <Bell className="h-4 w-4" />
        {unreadCount > 0 && (
          <span className="bg-danger text-danger-foreground animate-fade-in absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-bold">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </button>

      {open && (
        <div className="bg-popover border-border animate-fade-in absolute top-full right-0 z-50 mt-2 w-96 overflow-hidden rounded-lg border shadow-xl">
          {/* Header */}
          <div className="border-border flex items-center justify-between border-b px-4 py-3">
            <div className="flex items-center gap-2">
              <h3 className="text-h4 text-foreground">Notifications</h3>
              {unreadCount > 0 && (
                <span className="bg-danger text-danger-foreground text-caption inline-flex h-5 min-w-5 items-center justify-center rounded-full px-1.5 font-semibold">
                  {unreadCount}
                </span>
              )}
            </div>
            <div className="flex items-center gap-1">
              {unreadCount > 0 && (
                <NexusButton
                  variant="ghost"
                  size="sm"
                  onClick={markAllAsRead}
                  className="text-caption"
                >
                  <CheckCheck className="mr-1 h-3.5 w-3.5" />
                  Read all
                </NexusButton>
              )}
              <NexusButton
                variant="ghost"
                size="sm"
                onClick={clearAll}
                className="text-caption text-muted-foreground"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </NexusButton>
            </div>
          </div>

          {/* List */}
          <div className="divide-border max-h-96 divide-y overflow-y-auto">
            {notifications.length === 0 ? (
              <div className="text-muted-foreground flex flex-col items-center justify-center py-10">
                <Bell className="mb-2 h-8 w-8 opacity-30" />
                <p className="text-body">No notifications</p>
              </div>
            ) : (
              notifications.map((n) => (
                <button
                  key={n.id}
                  onClick={() => markAsRead(n.id)}
                  className={cn(
                    "hover:bg-surface-hover flex w-full items-start gap-3 px-4 py-3 text-left transition-colors",
                    !n.read && "bg-primary/5",
                  )}
                >
                  <span className="mt-0.5 shrink-0">{typeIcons[n.type]}</span>
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
                  <div className="flex shrink-0 flex-col items-end gap-1">
                    <span className="text-caption text-muted-foreground">
                      {n.time}
                    </span>
                    {!n.read && (
                      <span className="bg-primary h-2 w-2 rounded-full" />
                    )}
                  </div>
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
