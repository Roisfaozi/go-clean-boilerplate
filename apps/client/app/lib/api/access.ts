import { apiClient } from "./client";
import type { PaginatedResponse } from "./types";

export interface AccessRight {
	id: string;
	name: string;
	resource: string;
	action: string;
	conditions?: Record<string, any>;
	created_at: number;
	updated_at: number;
}

export interface Endpoint {
	id: string;
	name: string;
	method: string;
	path: string;
	resource_id: string;
	description?: string;
	created_at: number;
	updated_at: number;
}

export const accessApi = {
	listRights: (params?: { page?: number; limit?: number }) =>
		apiClient.get<PaginatedResponse<AccessRight>>("/access-rights", undefined, {
			params,
		}),

	searchRights: (params: any) =>
		apiClient.post<PaginatedResponse<AccessRight>>(
			"/access-rights/search",
			params,
		),

	listEndpoints: (params?: { page?: number; limit?: number }) =>
		apiClient.get<PaginatedResponse<Endpoint>>("/endpoints", undefined, {
			params,
		}),

	searchEndpoints: (params: any) =>
		apiClient.post<PaginatedResponse<Endpoint>>("/endpoints/search", params),

	link: (data: { access_right_id: string; endpoint_id: string }) =>
		apiClient.post("/access-rights/link", data),

	unlink: (data: { access_right_id: string; endpoint_id: string }) =>
		apiClient.post("/access-rights/unlink", data),
};
