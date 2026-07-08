import { apiClient } from "./client";
import {
	projectSchema,
	paginatedSchema,
	type Project,
	type PaginatedResponse,
} from "./schemas";

const paginatedProjects = paginatedSchema(projectSchema);

export const projectsApi = {
	list: (params?: { page?: number; limit?: number; org_id?: string }) =>
		apiClient.get<PaginatedResponse<Project>>("/projects", paginatedProjects, {
			params,
		}),

	create: (data: {
		name: string;
		slug?: string;
		description?: string;
		organization_id: string;
	}) => apiClient.post<Project>("/projects", data, projectSchema),

	update: (id: string, data: Partial<Project>) =>
		apiClient.put<Project>(`/projects/${id}`, data, projectSchema),

	delete: (id: string) => apiClient.delete(`/projects/${id}`),
};
