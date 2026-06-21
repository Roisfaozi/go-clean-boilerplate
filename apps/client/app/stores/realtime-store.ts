import { create } from "zustand";

/* ═══════ Notification Store ═══════ */
export interface RealtimeNotification {
  id: string;
  title: string;
  description?: string;
  type: "info" | "success" | "warning" | "danger" | "system";
  time: string;
  read: boolean;
}

interface NotificationState {
  notifications: RealtimeNotification[];
  unreadCount: number;
  addNotification: (
    n: Omit<RealtimeNotification, "id" | "time" | "read">,
  ) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  clearAll: () => void;
}

export const useNotificationStore = create<NotificationState>((set, get) => ({
  notifications: [
    {
      id: "demo-1",
      title: "User alice@nexus.dev signed up",
      type: "info",
      time: "2m ago",
      read: false,
    },
    {
      id: "demo-2",
      title: "Deployment v2.1.0 complete",
      description: "All services healthy",
      type: "success",
      time: "15m ago",
      read: false,
    },
    {
      id: "demo-3",
      title: "High error rate detected",
      description: "Error rate > 1% on /api/users",
      type: "danger",
      time: "1h ago",
      read: true,
    },
  ],
  unreadCount: 2,
  addNotification: (n) => {
    const notification: RealtimeNotification = {
      ...n,
      id: `notif-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
      time: "just now",
      read: false,
    };
    set((s) => ({
      notifications: [notification, ...s.notifications].slice(0, 50),
      unreadCount: s.unreadCount + 1,
    }));
  },
  markAsRead: (id) =>
    set((s) => {
      const notifications = s.notifications.map((n) =>
        n.id === id ? { ...n, read: true } : n,
      );
      return {
        notifications,
        unreadCount: notifications.filter((n) => !n.read).length,
      };
    }),
  markAllAsRead: () =>
    set((s) => ({
      notifications: s.notifications.map((n) => ({ ...n, read: true })),
      unreadCount: 0,
    })),
  clearAll: () => set({ notifications: [], unreadCount: 0 }),
}));

/* ═══════ Presence Store ═══════ */
export interface PresenceUser {
  id: string;
  name: string;
  avatar?: string;
  status: "online" | "away" | "busy";
  lastSeen?: string;
}

interface PresenceState {
  users: PresenceUser[];
  onlineCount: number;
  addUser: (user: PresenceUser) => void;
  removeUser: (id: string) => void;
  updateUser: (id: string, data: Partial<PresenceUser>) => void;
  setUsers: (users: PresenceUser[]) => void;
}

export const usePresenceStore = create<PresenceState>((set) => ({
  users: [
    { id: "u1", name: "Alice Chen", status: "online" },
    { id: "u2", name: "Bob Smith", status: "online" },
    { id: "u3", name: "Carol Davis", status: "away" },
    { id: "u4", name: "Dave Wilson", status: "online" },
    { id: "u5", name: "Eve Johnson", status: "busy" },
  ],
  onlineCount: 3,
  addUser: (user) =>
    set((s) => {
      if (s.users.find((u) => u.id === user.id)) return s;
      const users = [...s.users, user];
      return {
        users,
        onlineCount: users.filter((u) => u.status === "online").length,
      };
    }),
  removeUser: (id) =>
    set((s) => {
      const users = s.users.filter((u) => u.id !== id);
      return {
        users,
        onlineCount: users.filter((u) => u.status === "online").length,
      };
    }),
  updateUser: (id, data) =>
    set((s) => {
      const users = s.users.map((u) => (u.id === id ? { ...u, ...data } : u));
      return {
        users,
        onlineCount: users.filter((u) => u.status === "online").length,
      };
    }),
  setUsers: (users) =>
    set({
      users,
      onlineCount: users.filter((u) => u.status === "online").length,
    }),
}));

/* ═══════ Activity Feed Store ═══════ */
export interface LiveActivity {
  id: string;
  user: string;
  action: string;
  target?: string;
  time: string;
  type: "info" | "success" | "warning" | "danger";
}

interface ActivityState {
  activities: LiveActivity[];
  addActivity: (a: Omit<LiveActivity, "id" | "time">) => void;
  clearActivities: () => void;
}

export const useActivityStore = create<ActivityState>((set) => ({
  activities: [
    {
      id: "a1",
      user: "Alice",
      action: "deployed",
      target: "v2.1.0 to production",
      time: "2m ago",
      type: "success",
    },
    {
      id: "a2",
      user: "Bob",
      action: "updated role",
      target: "Editor permissions",
      time: "5m ago",
      type: "info",
    },
    {
      id: "a3",
      user: "System",
      action: "auto-scaled",
      target: "API servers to 4 instances",
      time: "12m ago",
      type: "warning",
    },
    {
      id: "a4",
      user: "Carol",
      action: "invited",
      target: "dave@nexus.dev",
      time: "18m ago",
      type: "info",
    },
    {
      id: "a5",
      user: "Eve",
      action: "deleted",
      target: "staging-old project",
      time: "25m ago",
      type: "danger",
    },
  ],
  addActivity: (a) => {
    const activity: LiveActivity = {
      ...a,
      id: `act-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
      time: "just now",
    };
    set((s) => ({
      activities: [activity, ...s.activities].slice(0, 100),
    }));
  },
  clearActivities: () => set({ activities: [] }),
}));

/* ═══════ Connection Status Store ═══════ */
interface ConnectionState {
  sseConnected: boolean;
  wsConnected: boolean;
  setSseConnected: (v: boolean) => void;
  setWsConnected: (v: boolean) => void;
}

export const useConnectionStore = create<ConnectionState>((set) => ({
  sseConnected: false,
  wsConnected: false,
  setSseConnected: (v) => set({ sseConnected: v }),
  setWsConnected: (v) => set({ wsConnected: v }),
}));

/* ═══════ Metrics Store ═══════ */
export interface SystemMetrics {
  rps: number;
  active_users: number;
  total_users: number;
  avg_latency: number;
  error_rate: number;
  uptime: number;
  cpu_usage: number;
  memory_usage: number;
  active_threads: number;
}

interface MetricsState {
  metrics: SystemMetrics | null;
  history: { time: string; rps: number; latency: number }[];
  updateMetrics: (m: SystemMetrics) => void;
}

export const useMetricsStore = create<MetricsState>((set) => ({
  metrics: null,
  history: [],
  updateMetrics: (m) =>
    set((s) => {
      const now = new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      });
      const newHistory = [
        ...s.history,
        { time: now, rps: m.rps, latency: m.avg_latency },
      ].slice(-20);
      return { metrics: m, history: newHistory };
    }),
}));
