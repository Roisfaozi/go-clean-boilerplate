"use client";

import { useEffect, useState } from "react";
import { useWebSocket } from "~/components/shared/providers/websocket-provider";
import { AuditLog } from "~/lib/api/audit";
import { toast } from "sonner";

export function useAuditStream() {
  const { subscribe, unsubscribe } = useWebSocket();
  const [newLog, setNewLog] = useState<AuditLog | null>(null);

  useEffect(() => {
    const handleNewLog = (message: any) => {
      // message format: { type: "message", channel: "audit", data: AuditLog }
      if (message.data) {
        setNewLog(message.data);
        toast.info(`New Audit Log: ${message.data.action}`, {
          description: `User ${message.data.user_id} on ${message.data.entity}`,
        });
      }
    };

    subscribe("audit", handleNewLog);

    return () => {
      unsubscribe("audit", handleNewLog);
    };
  }, [subscribe, unsubscribe]);

  return newLog;
}
