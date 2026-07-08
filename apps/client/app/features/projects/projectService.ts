import { apiClient } from "@/lib/api/client";
import type { Project, PaginatedResponse } from "@/lib/api/schemas";

export const projectService = {
	list: (params?: { page?: number; limit?: number; org_id?: string }) =>
		apiClient.get<PaginatedResponse<Project>>("/projects", undefined, {
			params,
		}),
	create: (data: {
		name: string;
		slug?: string;
		description?: string;
		organization_id: string;
	}) => apiClient.post<Project>("/projects", data),
	update: (id: string, data: Partial<Project>) =>
		apiClient.put<Project>(`/projects/${id}`, data),
	delete: (id: string) => apiClient.delete(`/projects/${id}`),
};
