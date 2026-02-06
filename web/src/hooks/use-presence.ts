"use client";

import { useEffect, useCallback } from "react";
import { useWebSocket } from "~/components/shared/providers/websocket-provider";
import { usePresenceStore, PresenceUser } from "~/stores/use-presence-store";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { organizationsApi } from "~/lib/api/organizations";

export function usePresence() {
  const { isConnected, subscribe, unsubscribe, sendJson } = useWebSocket();
  const { currentOrganization } = useOrganizationStore();
  const { setOnlineUsers, addUser, removeUser } = usePresenceStore();

  const handlePresenceUpdate = useCallback(
    (message: any) => {
      // Expected message format: { type: "message", channel: "presence:org:...", data: { event: "join|leave", user: {...} } }
      const { event, user } = message.data;
      if (event === "join") {
        addUser(user);
      } else if (event === "leave" || event === "timeout") {
        removeUser(user.user_id);
      }
    },
    [addUser, removeUser]
  );

  useEffect(() => {
    if (!currentOrganization || !isConnected) return;

    const channel = `presence:org:${currentOrganization.id}`;

    // 1. Fetch initial state
    organizationsApi.getPresence(currentOrganization.id).then((resp) => {
      if (resp.data) {
        // Map backend Member[] to PresenceUser[]
        const users: PresenceUser[] = resp.data.map((m: any) => ({
          user_id: m.user_id,
          name: m.user?.name,
          avatar_url: m.user?.avatar_url,
          role: m.role?.name,
          status: m.status,
          last_seen: m.joined_at, // Use joined_at as fallback for last_seen if needed
        }));
        setOnlineUsers(users);
      }
    });

    // 2. Subscribe to live updates
    subscribe(channel, handlePresenceUpdate);

    // 3. Heartbeat interval (30s)
    const heartbeat = setInterval(() => {
      sendJson({ type: "presence_heartbeat" });
    }, 30000);

    return () => {
      unsubscribe(channel, handlePresenceUpdate);
      clearInterval(heartbeat);
    };
  }, [
    currentOrganization,
    isConnected,
    subscribe,
    unsubscribe,
    sendJson,
    setOnlineUsers,
    handlePresenceUpdate,
  ]);
}
