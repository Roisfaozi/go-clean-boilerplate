import { apiClient } from "./client";
import {
  roleSchema,
  paginatedSchema,
  type Role,
  type PaginatedResponse,
} from "./schemas";

const paginatedRoles = paginatedSchema(roleSchema);

export const rolesApi = {
  list: (params?: { page?: number; limit?: number }) =>
    apiClient.get<PaginatedResponse<Role>>("/roles", paginatedRoles, {
      params,
    }),

  create: (data: { name: string; description?: string }) =>
    apiClient.post<Role>("/roles", data, roleSchema),

  update: (id: string, data: { name?: string; description?: string }) =>
    apiClient.put<Role>(`/roles/${id}`, data, roleSchema),

  delete: (id: string) => apiClient.delete(`/roles/${id}`),
};
