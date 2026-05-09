import { apiClient } from '@/lib/api/client';
import type { PaginatedResponse } from '@/lib/api/schemas';
import type { Resource } from '@/lib/api/types';

export const resourceService = {
  list: (params?: { page?: number; limit?: number }) =>
    apiClient.get<PaginatedResponse<Resource>>('/resources', undefined, { params }),
  create: (data: { name: string; slug: string; description?: string }) =>
    apiClient.post<Resource>('/resources', data),
  update: (id: string, data: Partial<Resource>) =>
    apiClient.put<Resource>(`/resources/${id}`, data),
  delete: (id: string) =>
    apiClient.delete(`/resources/${id}`),
};
