"use client";

import * as React from "react";
import { Bell, BellDot, X, Info, CheckCircle2, AlertTriangle, AlertCircle, Trash2 } from "lucide-react";
import { useNotificationStore, Notification } from "~/stores/use-notification-store";
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
  const { notifications, addNotification, markAllAsRead, clearAll, markAsRead } = useNotificationStore();
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

  const unreadCount = notifications.filter(n => !n.read).length;

  const getIcon = (type: Notification["type"]) => {
    switch (type) {
      case "success": return <CheckCircle2 className="h-4 w-4 text-emerald-500" />;
      case "warning": return <AlertTriangle className="h-4 w-4 text-amber-500" />;
      case "error": return <AlertCircle className="h-4 w-4 text-destructive" />;
      default: return <Info className="h-4 w-4 text-blue-500" />;
    }
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon" className="relative h-9 w-9">
          {unreadCount > 0 ? (
            <>
              <BellDot className="h-5 w-5 text-primary animate-pulse" />
              <Badge className="absolute -top-1 -right-1 h-4 w-4 flex items-center justify-center p-0 bg-primary text-[10px]">
                {unreadCount > 9 ? "9+" : unreadCount}
              </Badge>
            </>
          ) : (
            <Bell className="h-5 w-5 text-muted-foreground" />
          )}
          <span className="sr-only">Notifications</span>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="end">
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center gap-2">
            <h4 className="font-semibold text-sm">Notifications</h4>
            {unreadCount > 0 && (
              <Badge variant="secondary" className="h-5 text-[10px]">{unreadCount} unread</Badge>
            )}
          </div>
          <div className="flex gap-1">
            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={markAllAsRead} title="Mark all as read">
              <CheckCircle2 className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={clearAll} title="Clear all">
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        </div>
        <ScrollArea className="h-[350px]">
          {notifications.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="p-3 bg-muted rounded-full mb-3">
                <Bell className="h-6 w-6 text-muted-foreground/50" />
              </div>
              <p className="text-sm text-muted-foreground">No notifications yet</p>
            </div>
          ) : (
            <div className="flex flex-col">
              {notifications.map((n) => (
                <div 
                  key={n.id} 
                  className={cn(
                    "flex gap-3 p-4 border-b last:border-0 hover:bg-muted/50 transition-colors cursor-default relative",
                    !n.read && "bg-primary/5"
                  )}
                  onClick={() => markAsRead(n.id)}
                >
                  <div className="mt-0.5">{getIcon(n.type)}</div>
                  <div className="flex-1 space-y-1">
                    <p className={cn("text-xs font-semibold leading-none", !n.read && "text-primary")}>{n.title}</p>
                    <p className="text-[11px] text-muted-foreground line-clamp-2 leading-tight">
                      {n.description}
                    </p>
                    <p className="text-[10px] text-muted-foreground pt-1">
                      {formatDistanceToNow(n.createdAt, { addSuffix: true })}
                    </p>
                  </div>
                  {!n.read && (
                    <div className="absolute top-4 right-4 h-2 w-2 rounded-full bg-primary" />
                  )}
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
        <div className="p-2 border-t text-center">
            <Button variant="ghost" size="sm" className="w-full text-xs text-muted-foreground" onClick={() => setOpen(false)}>
                Close
            </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
