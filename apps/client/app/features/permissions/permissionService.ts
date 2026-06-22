import { apiClient } from "@/lib/api/client";
import type { Permission, PaginatedResponse } from "@/lib/api/schemas";

export const permissionService = {
  list: (params?: { page?: number; limit?: number; role_id?: string }) =>
    apiClient.get<PaginatedResponse<Permission>>("/permissions", undefined, {
      params,
    }),
  create: (data: {
    role_id: string;
    access_right_id: string;
    granted: boolean;
  }) => apiClient.post<Permission>("/permissions", data),
  update: (id: string, data: Partial<Permission>) =>
    apiClient.put<Permission>(`/permissions/${id}`, data),
  delete: (id: string) => apiClient.delete(`/permissions/${id}`),

  // Matrix View Endpoints
  getResources: () =>
    apiClient.get<{ resources: any[] }>("/permissions/resources"),
  getRoleAccessRights: (role: string, domain?: string) =>
    apiClient.get<any>(`/permissions/roles/${role}/access-rights`, undefined, {
      params: { domain },
    }),
  assignAccessRight: (data: {
    role: string;
    access_right_id: string;
    domain?: string;
  }) => apiClient.post("/permissions/assign-access-right", data),
  revokeAccessRight: (data: {
    role: string;
    access_right_id: string;
    domain?: string;
  }) => apiClient.delete("/permissions/revoke-access-right", { data }),

  // Inheritance Endpoints
  getInheritanceTree: () =>
    apiClient.get<{ roles: any[] }>("/permissions/inheritance-tree"),
  addInheritance: (data: {
    child_role: string;
    parent_role: string;
    domain?: string;
  }) => apiClient.post("/permissions/inheritance", data),
  removeInheritance: (data: {
    child_role: string;
    parent_role: string;
    domain?: string;
  }) => apiClient.delete("/permissions/inheritance", { data }),
};
