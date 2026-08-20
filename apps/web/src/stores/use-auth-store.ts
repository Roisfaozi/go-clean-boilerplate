import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface User {
	id: string;
	name: string;
	email: string;
	username: string;
	role: string;
	avatar_url?: string;
}

interface AuthState {
	user: User | null;
	hasHydrated: boolean;
	setUser: (user: User | null) => void;
	setHasHydrated: (hydrated: boolean) => void;
	logout: () => void;
}

export const useAuthStore = create<AuthState>()(
	persist(
		(set) => ({
			user: null,
			hasHydrated: false,
			setUser: (user) => set({ user }),
			setHasHydrated: (hydrated) => set({ hasHydrated: hydrated }),
			logout: () => set({ user: null }),
		}),
		{
			name: "nexus-auth-storage",
			onRehydrateStorage: () => (state) => {
				state?.setHasHydrated(true);
			},
		},
	),
);
