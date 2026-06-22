import { apiClient } from "@/lib/api/client";
import type { PaginatedResponse } from "@/lib/api/schemas";
import type { Endpoint } from "@/lib/api/types";

export const endpointService = {
  list: (params?: any) =>
    apiClient.post<PaginatedResponse<Endpoint>>("/endpoints/search", params || {}),
  create: (data: {
    name: string;
    method: string;
    path: string;
    resource_id: string;
    description?: string;
    auth_required?: boolean;
  }) => apiClient.post<Endpoint>("/endpoints", data),
  update: (id: string, data: Partial<Endpoint>) =>
    apiClient.put<Endpoint>(`/endpoints/${id}`, data),
  delete: (id: string) => apiClient.delete(`/endpoints/${id}`),
};
