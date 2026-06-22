import { apiClient } from "@/lib/api/client";
import type { AccessRight, PaginatedResponse } from "@/lib/api/schemas";

export const accessRightService = {
  list: (params?: { page?: number; limit?: number }) =>
    apiClient.get<PaginatedResponse<AccessRight>>("/access-rights", undefined, {
      params,
    }),
  create: (data: {
    name: string;
    resource: string;
    action: string;
    conditions?: Record<string, unknown>;
  }) => apiClient.post<AccessRight>("/access-rights", data),
  update: (id: string, data: Partial<AccessRight>) =>
    apiClient.put<AccessRight>(`/access-rights/${id}`, data),
  delete: (id: string) => apiClient.delete(`/access-rights/${id}`),
};
