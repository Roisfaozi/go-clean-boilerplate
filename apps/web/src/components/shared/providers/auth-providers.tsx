"use client";

import { usePathname } from "next/navigation";
import type React from "react";
import { useEffect } from "react";
import { accessApi } from "~/lib/api/access";
import { authApi } from "~/lib/api/auth";
import { useAuthStore } from "~/stores/use-auth-store";
import { usePermissionStore } from "~/stores/use-permission-store";

const AUTH_PATHS = [
	"/login",
	"/register",
	"/forgot-password",
	"/reset-password",
];

export function AuthProvider({ children }: { children: React.ReactNode }) {
	const { hasHydrated, setUser, logout } = useAuthStore();
	const { setPermissions, clearPermissions } = usePermissionStore();
	const pathname = usePathname();

	const isAuthPage = AUTH_PATHS.some((p) => pathname?.includes(p));

	useEffect(() => {
		if (!hasHydrated || isAuthPage) return;

		let cancelled = false;

		async function syncAuth() {
			try {
				const userResp = await authApi.getCurrentUser();
				const user = userResp?.data?.user;
				if (user) {
					if (cancelled) return;
					setUser(user);

					const permsResp = await accessApi.getPermissionsForRole(user.role);
					if (!cancelled && permsResp.data) {
						setPermissions(permsResp.data);
					}
				} else {
					logout();
					clearPermissions();
				}
			} catch (error) {
				console.error("Auth sync failed:", error);
				logout();
				clearPermissions();
			}
		}

		syncAuth();

		return () => {
			cancelled = true;
		};
	}, [
		hasHydrated,
		isAuthPage,
		setUser,
		logout,
		setPermissions,
		clearPermissions,
	]);

	return <>{children}</>;
}
