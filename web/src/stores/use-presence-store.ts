import { create } from "zustand";

export interface PresenceUser {
  user_id: string;
  name?: string;
  avatar_url?: string;
  role?: string;
  status: string;
  last_seen: number;
}

interface PresenceState {
  onlineUsers: PresenceUser[];
  setOnlineUsers: (users: PresenceUser[]) => void;
  addUser: (user: PresenceUser) => void;
  removeUser: (userId: string) => void;
}

export const usePresenceStore = create<PresenceState>((set) => ({
  onlineUsers: [],
  setOnlineUsers: (users) => set({ onlineUsers: users }),
  addUser: (user) =>
    set((state) => {
      // Check if already in list to avoid duplicates from multiple tabs
      const exists = state.onlineUsers.some((u) => u.user_id === user.user_id);
      if (exists) return state;
      return { onlineUsers: [...state.onlineUsers, user] };
    }),
  removeUser: (userId) =>
    set((state) => ({
      onlineUsers: state.onlineUsers.filter((u) => u.user_id !== userId),
    })),
}));
