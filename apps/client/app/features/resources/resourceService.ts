import { apiClient } from "@/lib/api/client";
import type { PaginatedResponse } from "@/lib/api/schemas";
import type { Resource } from "@/lib/api/types";

export const resourceService = {
  list: () => apiClient.get<PaginatedResponse<Resource>>("/access-rights"),
  create: (data: { name: string; description?: string }) =>
    apiClient.post<Resource>("/access-rights", data),
  update: (id: string, data: Partial<Resource>) =>
    apiClient.put<Resource>(`/access-rights/${id}`, data),
  delete: (id: string) => apiClient.delete(`/access-rights/${id}`),
};
