"use client";

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { authApi } from "~/lib/api/auth";
import { useAuthStore } from "~/stores/use-auth-store";
import { useOrganizationStore } from "~/stores/use-organization-store";

interface WebSocketContextType {
  isConnected: boolean;
  subscribe: (channel: string, callback: (data: any) => void) => void;
  unsubscribe: (channel: string, callback: (data: any) => void) => void;
  sendJson: (data: any) => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";
const RECONNECT_INTERVAL = 5000;

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const [isConnected, setIsConnected] = useState(false);
  const { currentOrganization } = useOrganizationStore();
  const { user } = useAuthStore();
  const socketRef = useRef<WebSocket | null>(null);
  const subscriptions = useRef<Map<string, Set<(data: any) => void>>>(
    new Map()
  );
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const connectRef = useRef<() => void>(() => {});

  const sendJson = useCallback((data: any) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify(data));
    }
  }, []);

  const connect = useCallback(async () => {
    if (socketRef.current?.readyState === WebSocket.OPEN) return;
    if (!currentOrganization?.id || !user) return;

    try {
      // 1. Fetch short-lived ticket via Proxy (Secure HTTP context)
      const ticketData = await authApi.getWsTicket(currentOrganization?.id);
      const ticket = ticketData.ticket;

      // 2. Connect using the ticket
      const socket = new WebSocket(`${WS_URL}?ticket=${ticket}`);
      socketRef.current = socket;

      socket.onopen = () => {
        console.log("WebSocket connected with ticket");
        setIsConnected(true);
        // Resubscribe to channels if any (after reconnect)
        subscriptions.current.forEach((_, channel) => {
          sendJson({ type: "subscribe", channel });
        });
      };

      socket.onmessage = (event) => {
        try {
          // The backend might send multiple JSON objects separated by newlines in one frame
          const rawData = event.data as string;
          const lines = rawData
            .split("\n")
            .filter((line) => line.trim() !== "");

          for (const line of lines) {
            try {
              const message = JSON.parse(line);
              const channel = message.channel || "global";
              const listeners = subscriptions.current.get(channel);
              if (listeners) {
                listeners.forEach((callback) => callback(message));
              }
            } catch (parseError) {
              console.error(
                "Failed to parse WS message line:",
                line,
                parseError
              );
            }
          }
        } catch (error) {
          console.error("Failed to process WS message", error);
        }
      };

      socket.onclose = (event) => {
        console.log("WebSocket disconnected", event.reason);
        setIsConnected(false);
        socketRef.current = null;

        // Don't reconnect if it was a normal closure or if there was an auth error (4001 custom code)
        if (event.code === 1000 || event.code === 4001) return;

        // Reconnect logic
        if (reconnectTimeoutRef.current)
          clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = setTimeout(
          () => connectRef.current(),
          RECONNECT_INTERVAL
        );
      };

      socket.onerror = (error) => {
        console.error("WebSocket error:", error);
      };
    } catch (error) {
      console.error("Failed to establish WebSocket connection:", error);

      const isAuthError =
        error instanceof Error &&
        (error.message.includes("401") ||
          error.message.toLowerCase().includes("unauthorized") ||
          error.message.toLowerCase().includes("token"));

      if (isAuthError) return;

      if (reconnectTimeoutRef.current)
        clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = setTimeout(
        () => connectRef.current(),
        RECONNECT_INTERVAL
      );
    }
  }, [sendJson, currentOrganization, user]);

  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  // Handle connection/reconnection based on organization context
  useEffect(() => {
    if (socketRef.current) {
      socketRef.current.close(1000, "Organization Context Changed");
    }
    connectRef.current();
  }, [currentOrganization?.id]);

  useEffect(() => {
    return () => {
      if (socketRef.current) {
        socketRef.current.close(1000, "Component Unmounted");
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, []);

  const subscribe = useCallback(
    (channel: string, callback: (data: any) => void) => {
      if (!subscriptions.current.has(channel)) {
        subscriptions.current.set(channel, new Set());
        sendJson({ type: "subscribe", channel });
      }
      subscriptions.current.get(channel)?.add(callback);
    },
    [sendJson]
  );

  const unsubscribe = useCallback(
    (channel: string, callback: (data: any) => void) => {
      const listeners = subscriptions.current.get(channel);
      if (listeners) {
        listeners.delete(callback);
        if (listeners.size === 0) {
          subscriptions.current.delete(channel);
          sendJson({ type: "unsubscribe", channel });
        }
      }
    },
    [sendJson]
  );

  return (
    <WebSocketContext.Provider
      value={{ isConnected, subscribe, unsubscribe, sendJson }}
    >
      {children}
    </WebSocketContext.Provider>
  );
}

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return context;
};
