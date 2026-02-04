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
        // Message format: { type: "message", channel: "audit", data: {...} }
        // Or sometimes just { channel: "audit", ... } depending on backend implementation
        // Our backend seems to send raw bytes from Redis or JSON.
        // Let's assume standard format from ws_client.go: BroadcastMessage

        // Wait, ws_client.go WritePump sends raw bytes from 'Send' channel.
        // ws_manager broadcasts []byte.
        // The broadcast message from backend audit usecase is just the Log JSON.
        // But ws_manager doesn't wrap it in envelope?
        // Checking ws_client.go again...
        // WritePump just writes whatever is in Send channel.
        // Manager sends msg.Message directly.
        // So the message received here is whatever UseCase marshaled.

        // ISSUE: If multiple channels exist, how do we know which channel it belongs to?
        // ws_manager.go `handleBroadcast` -> sends `msg.Message` to `client.Send`.
        // It DOES NOT wrap it with channel name.
        // This is a flaw in the backend implementation if we want multiplexing on a single connection.
        // Ideally, the backend should wrap it: { channel: "audit", data: ... }

        // For now, let's assume the data itself might contain a "type" or we infer it,
        // OR we fix the backend to wrap messages.
        // Fixing backend is safer.

        // Let's implement the frontend assuming the backend WILL send:
        // { channel: "audit", payload: ... }

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
      reconnectTimeoutRef.current = setTimeout(connect, RECONNECT_INTERVAL);
    };

    socket.onerror = (error) => {
      console.error("WebSocket error", error);
      socket.close();
    };
  }, []);

  useEffect(() => {
    connect();
    return () => {
      if (socketRef.current) {
        socketRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [connect]);

  const sendJson = (data: any) => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify(data));
    }
  };

  const subscribe = useCallback(
    (channel: string, callback: (data: any) => void) => {
      if (!subscriptions.current.has(channel)) {
        subscriptions.current.set(channel, new Set());
        // Send subscribe request to server
        sendJson({ type: "subscribe", channel });
      }
      subscriptions.current.get(channel)?.add(callback);
    },
    []
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
    []
  );

  return (
    <WebSocketContext.Provider value={{ isConnected, subscribe, unsubscribe }}>
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
