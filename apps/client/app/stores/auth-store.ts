import type { User } from '@/lib/api/schemas'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  initialized: boolean
  login: (user: User, token?: string) => void
  logout: () => void
  setUser: (user: User) => void
  setToken: (token: string) => void
  setInitialized: (initialized: boolean) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token:
        typeof window !== 'undefined'
          ? localStorage.getItem('nexus_token')
          : null,
      isAuthenticated: false,
      initialized: false,
      login: (user, token) => {
        if (token) {
          localStorage.setItem('nexus_token', token)
        }
        set({ user, token: token ?? null, isAuthenticated: true })
      },
      logout: () => {
        if (typeof window !== 'undefined') {
          localStorage.removeItem('nexus_token')
        }
        set({
          user: null,
          token: null,
          isAuthenticated: false,
          initialized: true,
        })
      },
      setUser: (user) => set({ user, isAuthenticated: true }),
      setToken: (token) => {
        if (typeof window !== 'undefined') {
          localStorage.setItem('nexus_token', token)
        }
        set({ token, isAuthenticated: true })
      },
      setInitialized: (initialized) => set({ initialized }),
    }),
    { name: 'nexus-auth' },
  ),
)
