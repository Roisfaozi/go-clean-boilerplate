"use client";

import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  useCallback,
} from "react";
import { toast } from "sonner";

interface WebSocketContextType {
  isConnected: boolean;
  subscribe: (channel: string, callback: (data: any) => void) => void;
  unsubscribe: (channel: string, callback: (data: any) => void) => void;
  sendJson: (data: any) => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";
const RECONNECT_INTERVAL = 3000;

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const [isConnected, setIsConnected] = useState(false);
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

  const connect = useCallback(() => {
    if (socketRef.current?.readyState === WebSocket.OPEN) return;

    // Use built-in WebSocket (cookie authentication is automatic for same-origin/configured CORS)
    const socket = new WebSocket(WS_URL);
    socketRef.current = socket;

    socket.onopen = () => {
      console.log("WebSocket connected");
      setIsConnected(true);
      // Resubscribe to channels if any (after reconnect)
      subscriptions.current.forEach((_, channel) => {
        sendJson({ type: "subscribe", channel });
      });
    };

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        const channel = message.channel || "global";
        const listeners = subscriptions.current.get(channel);
        if (listeners) {
          listeners.forEach((callback) => callback(message));
        }
      } catch (error) {
        console.error("Failed to parse WS message", error);
      }
    };

    socket.onclose = () => {
      console.log("WebSocket disconnected");
      setIsConnected(false);
      socketRef.current = null;
      // Reconnect logic
      reconnectTimeoutRef.current = setTimeout(
        () => connectRef.current(),
        RECONNECT_INTERVAL
      );
    };

    socket.onerror = (error) => {
      console.error("WebSocket error", error);
      socket.close();
    };
  }, [sendJson]);

  useEffect(() => {
    connectRef.current = connect;
  }, [connect]);

  useEffect(() => {
    connectRef.current();
    return () => {
      if (socketRef.current) {
        socketRef.current.close();
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
        // Send subscribe request to server
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
