"use client";

import { usePermissionStore } from "~/stores/use-permission-store";

/**
 * Hook to check if the current user has permission for a specific resource and action.
 *
 * @param resource The resource path (e.g., "/api/v1/users")
 * @param action The HTTP method or action (e.g., "GET", "POST", "DELETE")
 * @returns boolean
 */
export function usePermission(resource?: string, action?: string) {
  const hasPermission = usePermissionStore((state) => state.hasPermission);

  if (!resource || !action) {
    return false;
  }

  return hasPermission(resource, action);
}

/**
 * Hook to check multiple permissions.
 * Useful for checking if user can see a whole section.
 */
export function usePermissions(items: { resource: string; action: string }[]) {
  const hasPermission = usePermissionStore((state) => state.hasPermission);

  return items.every((item) => hasPermission(item.resource, item.action));
}
