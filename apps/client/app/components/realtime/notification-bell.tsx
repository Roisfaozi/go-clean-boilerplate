import { useState, useRef, useEffect } from "react";
import { cn } from "@casbin/ui";
import { useNotificationStore, type RealtimeNotification } from "@/stores/realtime-store";
import { Bell, Check, CheckCheck, Trash2, Info, CheckCircle2, AlertTriangle, AlertCircle, Radio } from "lucide-react";
import {  NexusButton  } from "@casbin/ui";

const typeIcons: Record<RealtimeNotification["type"], React.ReactNode> = {
  info: <Info className="h-4 w-4 text-info" />,
  success: <CheckCircle2 className="h-4 w-4 text-success" />,
  warning: <AlertTriangle className="h-4 w-4 text-warning" />,
  danger: <AlertCircle className="h-4 w-4 text-danger" />,
  system: <Radio className="h-4 w-4 text-primary" />,
};

interface NotificationBellProps {
  className?: string;
}

export function NotificationBell({ className }: NotificationBellProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const { notifications, unreadCount, markAsRead, markAllAsRead, clearAll } = useNotificationStore();

  // Close on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  return (
    <div ref={ref} className={cn("relative", className)}>
      <button
        onClick={() => setOpen(!open)}
        className="relative p-2 rounded-md hover:bg-surface-hover text-muted-foreground transition-colors"
      >
        <Bell className="h-4 w-4" />
        {unreadCount > 0 && (
          <span className="absolute -top-0.5 -right-0.5 flex items-center justify-center h-4 min-w-4 rounded-full bg-danger text-danger-foreground text-[10px] font-bold px-1 animate-fade-in">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-2 w-96 bg-popover border border-border rounded-lg shadow-xl z-50 animate-fade-in overflow-hidden">
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-border">
            <div className="flex items-center gap-2">
              <h3 className="text-h4 text-foreground">Notifications</h3>
              {unreadCount > 0 && (
                <span className="inline-flex items-center justify-center h-5 min-w-5 rounded-full bg-danger text-danger-foreground text-caption px-1.5 font-semibold">
                  {unreadCount}
                </span>
              )}
            </div>
            <div className="flex items-center gap-1">
              {unreadCount > 0 && (
                <NexusButton variant="ghost" size="sm" onClick={markAllAsRead} className="text-caption">
                  <CheckCheck className="h-3.5 w-3.5 mr-1" />
                  Read all
                </NexusButton>
              )}
              <NexusButton variant="ghost" size="sm" onClick={clearAll} className="text-caption text-muted-foreground">
                <Trash2 className="h-3.5 w-3.5" />
              </NexusButton>
            </div>
          </div>

          {/* List */}
          <div className="max-h-96 overflow-y-auto divide-y divide-border">
            {notifications.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-10 text-muted-foreground">
                <Bell className="h-8 w-8 mb-2 opacity-30" />
                <p className="text-body">No notifications</p>
              </div>
            ) : (
              notifications.map((n) => (
                <button
                  key={n.id}
                  onClick={() => markAsRead(n.id)}
                  className={cn(
                    "flex items-start gap-3 w-full px-4 py-3 text-left hover:bg-surface-hover transition-colors",
                    !n.read && "bg-primary/5"
                  )}
                >
                  <span className="mt-0.5 shrink-0">{typeIcons[n.type]}</span>
                  <div className="flex-1 min-w-0">
                    <p className={cn("text-body truncate", !n.read ? "font-semibold text-foreground" : "text-foreground")}>
                      {n.title}
                    </p>
                    {n.description && <p className="text-caption text-muted-foreground truncate">{n.description}</p>}
                  </div>
                  <div className="flex flex-col items-end gap-1 shrink-0">
                    <span className="text-caption text-muted-foreground">{n.time}</span>
                    {!n.read && <span className="h-2 w-2 rounded-full bg-primary" />}
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
