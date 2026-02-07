"use client";

import React, { useEffect } from "react";
import { useAuthStore } from "~/stores/use-auth-store";
import { usePermissionStore } from "~/stores/use-permission-store";
import { authApi } from "~/lib/api/auth";
import { accessApi } from "~/lib/api/access";

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { user, setUser, logout } = useAuthStore();
  const { setPermissions, clearPermissions } = usePermissionStore();

  useEffect(() => {
    async function syncAuth() {
      try {
        // 1. Fetch current user from /auth/me via Proxy
        const userResp = await authApi.getCurrentUser();
        if (userResp.user) {
          setUser(userResp.user);
          
          // 2. Sync permissions for this user's role
          const permsResp = await accessApi.getPermissionsForRole(userResp.user.role);
          if (permsResp.data) {
            setPermissions(permsResp.data);
          }
        } else {
          // No user found, clear state
          logout();
          clearPermissions();
        }
      } catch (error) {
        console.error("Auth sync failed:", error);
        // On error (e.g. 401), we clear local state
        logout();
        clearPermissions();
      }
    }

    syncAuth();
  }, [setUser, logout, setPermissions, clearPermissions]);

  return <>{children}</>;
}
