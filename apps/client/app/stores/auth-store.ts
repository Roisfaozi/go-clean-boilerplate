import type { User } from "@/lib/api/schemas";
import { create } from "zustand";
import { persist } from "zustand/middleware";

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  initialized: boolean;
  login: (user: User) => void;
  logout: () => void;
  setUser: (user: User) => void;
  setInitialized: (initialized: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      initialized: false,
      login: (user) => {
        set({ user, isAuthenticated: true });
      },
      logout: () => {
        set({
          user: null,
          isAuthenticated: false,
          initialized: true,
        });
      },
      setUser: (user) => set({ user, isAuthenticated: true }),
      setInitialized: (initialized) => set({ initialized }),
    }),
    { name: "nexus-auth" },
  ),
);
