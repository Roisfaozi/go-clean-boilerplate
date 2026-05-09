"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

interface PermissionState {
  permissions: string[][]; // Format: [[role, dom, obj, act], ...]
  setPermissions: (permissions: string[][]) => void;
  hasPermission: (resource: string, action: string) => boolean;
  clearPermissions: () => void;
}

export const usePermissionStore = create<PermissionState>()(
  persist(
    (set, get) => ({
      permissions: [],
      setPermissions: (permissions) => set({ permissions }),
      hasPermission: (resource: string, action: string) => {
        const { permissions } = get();
        // Casbin model: p = sub, dom, obj, act
        // Kita cek apakah ada rule yang cocok dengan resource (obj) dan action (act)
        // User bisa memiliki multiple roles, jadi kita cek semua rule yang ada
        return permissions.some(
          (p) =>
            p.length >= 4 &&
            (p[2] === resource || p[2] === "*") &&
            (p[3] === action || p[3] === "*")
        );
      },
      clearPermissions: () => set({ permissions: [] }),
    }),
    {
      name: "nexus-permission-storage",
    }
  )
);
