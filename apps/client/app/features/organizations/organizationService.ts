import { apiClient } from "@/lib/api/client";
import type { Organization, PaginatedResponse } from "@/lib/api/schemas";

export const organizationService = {
  list: (params?: { page?: number; limit?: number }) =>
    apiClient.get<PaginatedResponse<Organization>>(
      "/organizations/me",
      undefined,
      { params },
    ),
  create: (data: { name: string; slug?: string }) =>
    apiClient.post<Organization>("/organizations", data),
  update: (id: string, data: Partial<Organization>) =>
    apiClient.put<Organization>(`/organizations/${id}`, data),
  delete: (id: string) => apiClient.delete(`/organizations/${id}`),
};
