"use client";

import * as React from "react";
import {
  Bell,
  BellDot,
  X,
  Info,
  CheckCircle2,
  AlertTriangle,
  AlertCircle,
  Trash2,
} from "lucide-react";
import {
  useNotificationStore,
  Notification,
} from "~/stores/use-notification-store";
import { useAuditStream } from "~/hooks/use-audit-stream";
import { Button } from "~/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "~/components/ui/popover";
import { ScrollArea } from "~/components/ui/scroll-area";
import { Badge } from "~/components/ui/badge";
import { cn } from "~/lib/utils";
import { formatDistanceToNow } from "date-fns";

export function NotificationCenter() {
  const {
    notifications,
    addNotification,
    markAllAsRead,
    clearAll,
    markAsRead,
  } = useNotificationStore();
  const newLog = useAuditStream();
  const [open, setOpen] = React.useState(false);

  // Auto-add notification when audit log arrives
  React.useEffect(() => {
    if (newLog) {
      addNotification({
        title: `System Action: ${newLog.action}`,
        description: `Entity ${newLog.entity} modified by user ${newLog.user_id}`,
        type: "info",
      });
    }
  }, [newLog, addNotification]);

  const unreadCount = notifications.filter((n) => !n.read).length;

  const getIcon = (type: Notification["type"]) => {
    switch (type) {
      case "success":
        return <CheckCircle2 className="h-4 w-4 text-emerald-500" />;
      case "warning":
        return <AlertTriangle className="h-4 w-4 text-amber-500" />;
      case "error":
        return <AlertCircle className="text-destructive h-4 w-4" />;
      default:
        return <Info className="h-4 w-4 text-blue-500" />;
    }
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative h-9 w-9">
          {unreadCount > 0 ? (
            <>
              <BellDot className="text-primary h-5 w-5 animate-pulse" />
              <Badge className="bg-primary absolute -top-1 -right-1 flex h-4 w-4 items-center justify-center p-0 text-[10px]">
                {unreadCount > 9 ? "9+" : unreadCount}
              </Badge>
            </>
          ) : (
            <Bell className="text-muted-foreground h-5 w-5" />
          )}
          <span className="sr-only">Notifications</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="end">
        <div className="flex items-center justify-between border-b p-4">
          <div className="flex items-center gap-2">
            <h4 className="text-sm font-semibold">Notifications</h4>
            {unreadCount > 0 && (
              <Badge variant="secondary" className="h-5 text-[10px]">
                {unreadCount} unread
              </Badge>
            )}
          </div>
          <div className="flex gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={markAllAsRead}
              title="Mark all as read"
            >
              <CheckCircle2 className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="text-destructive h-7 w-7"
              onClick={clearAll}
              title="Clear all"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        </div>
        <ScrollArea className="h-[350px]">
          {notifications.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="bg-muted mb-3 rounded-full p-3">
                <Bell className="text-muted-foreground/50 h-6 w-6" />
              </div>
              <p className="text-muted-foreground text-sm">
                No notifications yet
              </p>
            </div>
          ) : (
            <div className="flex flex-col">
              {notifications.map((n) => (
                <div
                  key={n.id}
                  className={cn(
                    "hover:bg-muted/50 relative flex cursor-default gap-3 border-b p-4 transition-colors last:border-0",
                    !n.read && "bg-primary/5"
                  )}
                  onClick={() => markAsRead(n.id)}
                >
                  <div className="mt-0.5">{getIcon(n.type)}</div>
                  <div className="flex-1 space-y-1">
                    <p
                      className={cn(
                        "text-xs leading-none font-semibold",
                        !n.read && "text-primary"
                      )}
                    >
                      {n.title}
                    </p>
                    <p className="text-muted-foreground line-clamp-2 text-[11px] leading-tight">
                      {n.description}
                    </p>
                    <p className="text-muted-foreground pt-1 text-[10px]">
                      {formatDistanceToNow(n.createdAt, { addSuffix: true })}
                    </p>
                  </div>
                  {!n.read && (
                    <div className="bg-primary absolute top-4 right-4 h-2 w-2 rounded-full" />
                  )}
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
        <div className="border-t p-2 text-center">
          <Button
            variant="ghost"
            size="sm"
            className="text-muted-foreground w-full text-xs"
            onClick={() => setOpen(false)}
          >
            Close
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
