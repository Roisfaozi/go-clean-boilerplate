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

const getLoginPath = (pathname: string) => {
	const localeMatch = pathname.match(/^\/([a-z]{2})(?:\/|$)/);
	const localePrefix = localeMatch ? `/${localeMatch[1]}` : "";
	const returnTo = encodeURIComponent(pathname);
	return `${localePrefix}/login?returnTo=${returnTo}`;
};

const hardLogout = async (pathname: string) => {
	await fetch("/api/auth/logout", {
		method: "POST",
		credentials: "include",
	}).catch(() => undefined);

	if (pathname.includes("/dashboard")) {
		window.location.href = getLoginPath(pathname);
	}
};

export function AuthProvider({ children }: { children: React.ReactNode }) {
	const { hasHydrated, setUser, logout } = useAuthStore();
	const { setPermissions, clearPermissions } = usePermissionStore();
	const pathname = usePathname() ?? "/";

	const isAuthPage = AUTH_PATHS.some((path) => pathname.includes(path));

	useEffect(() => {
		if (!hasHydrated || isAuthPage) return;

		let cancelled = false;

		async function syncAuth() {
			try {
				const userResp = await authApi.getCurrentUser();
				const user = userResp?.data?.user;

				if (!user) {
					logout();
					clearPermissions();
					await hardLogout(pathname);
					return;
				}

				if (cancelled) return;
				setUser(user);

				const permsResp = await accessApi.getPermissionsForRole(user.role);
				if (!cancelled && permsResp.data) {
					setPermissions(permsResp.data);
				}
			} catch (error) {
				if (cancelled) return;
				console.log("Auth Error", error);
				logout();
				clearPermissions();
				await hardLogout(pathname);
			}
		}

		syncAuth();

		return () => {
			cancelled = true;
		};
	}, [
		hasHydrated,
		isAuthPage,
		pathname,
		setUser,
		logout,
		setPermissions,
		clearPermissions,
	]);

	return <>{children}</>;
}
