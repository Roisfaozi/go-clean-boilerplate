import { apiClient } from "@/lib/api/client";
import type { Role, PaginatedResponse } from "@/lib/api/schemas";

export const roleService = {
	list: (params?: { page?: number; limit?: number }) =>
		apiClient.get<PaginatedResponse<Role>>("/roles", undefined, { params }),
	create: (data: { name: string; description?: string }) =>
		apiClient.post<Role>("/roles", data),
	update: (id: string, data: Partial<Role>) =>
		apiClient.put<Role>(`/roles/${id}`, data),
	delete: (id: string) => apiClient.delete(`/roles/${id}`),
};
