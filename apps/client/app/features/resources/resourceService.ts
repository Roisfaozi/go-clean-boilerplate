import { apiClient } from "@/lib/api/client";
import type { PaginatedResponse } from "@/lib/api/schemas";
import type { AccessRight } from "@/lib/api/types";

export const resourceService = {
  list: () => apiClient.get<PaginatedResponse<AccessRight>>("/access-rights"),
  create: (data: { name: string; resource: string; action: string }) =>
    apiClient.post<AccessRight>("/access-rights", data),
  update: (id: string, data: Partial<AccessRight>) =>
    apiClient.put<AccessRight>(`/access-rights/${id}`, data),
  delete: (id: string) => apiClient.delete(`/access-rights/${id}`),
};
