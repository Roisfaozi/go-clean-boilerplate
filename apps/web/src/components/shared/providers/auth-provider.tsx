"use client";

import { usePathname } from "next/navigation";
import type React from "react";
import { useEffect } from "react";
import { accessApi } from "~/lib/api/access";
import { authApi } from "~/lib/api/auth";
import { useAuthStore } from "~/stores/use-auth-store";
import { usePermissionStore } from "~/stores/use-permission-store";

/** Halaman auth: tidak perlu sync sama sekali */
const AUTH_PATHS = [
  "/login",
  "/register",
  "/forgot-password",
  "/reset-password",
];

/** Halaman publik: coba sync tapi JANGAN redirect jika gagal */
const PUBLIC_PATHS = ["/", "/about", "/changelog", "/pricing"];

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { setUser, logout } = useAuthStore();
  const { setPermissions, clearPermissions } = usePermissionStore();
  const pathname = usePathname();

  const isAuthPage = AUTH_PATHS.some((p) => pathname?.includes(p));
  const _isPublicPage =
    PUBLIC_PATHS.some((p) => pathname === p || pathname?.startsWith(p + "/")) ||
    (!pathname?.includes("/dashboard") && !isAuthPage);

  useEffect(() => {
    // Di halaman auth, tidak perlu sync
    if (isAuthPage) return;

    async function syncAuth() {
      try {
        const userResp = await authApi.getCurrentUser();
        if (userResp.user) {
          setUser(userResp.user);

          const permsResp = await accessApi.getPermissionsForRole(
            userResp.user.role,
          );
          if (permsResp.data) {
            setPermissions(permsResp.data);
          }
        } else {
          // Tidak ada user: hapus state lokal
          logout();
          clearPermissions();
          // Redirect ke login hanya dari area dashboard
          // (redirect dari public page ditangani oleh proxy.ts middleware)
        }
      } catch (error) {
        // Jika gagal (misal: 401 dari public page), bersihkan state saja
        // tanpa redirect — redirect dari public pages hanya terjadi jika
        // user secara aktif mencoba akses area yang protected (/dashboard)
        console.log("Auth Error", error);
        logout();
        clearPermissions();
      }
    }

    syncAuth();
  }, [isAuthPage, setUser, logout, setPermissions, clearPermissions]);

  return <>{children}</>;
}
