"use client";

import { usePathname } from "next/navigation";
import React, { useEffect } from "react";
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
  const { user, setUser, logout } = useAuthStore();
  const { setPermissions, clearPermissions } = usePermissionStore();
  const pathname = usePathname();

  const isAuthPage = AUTH_PATHS.some((p) => pathname?.includes(p));

  useEffect(() => {
    if (isAuthPage) return;

    async function syncAuth() {
      try {
        const userResp = await authApi.getCurrentUser();
        if (userResp.user) {
          setUser(userResp.user);

          const permsResp = await accessApi.getPermissionsForRole(
            userResp.user.role
          );
          if (permsResp.data) {
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
  }, [isAuthPage, setUser, logout, setPermissions, clearPermissions]);

  return <>{children}</>;
}
