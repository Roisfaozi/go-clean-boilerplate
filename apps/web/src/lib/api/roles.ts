import { api } from "./client";

export interface Role {
	id: string;
	name: string;
	description: string;
	organization_id?: string;
}

export interface RoleListResponse {
	data: Role[];
	paging: {
		limit: number;
		page: number;
		size: number;
		total: number;
		total_item: number;
		total_page: number;
	};
}

export interface SearchRoleFilter {
	page?: number;
	page_size?: number;
	filter?: Record<string, { type: string; [key: string]: any }>;
	sort?: { colId: string; sort: "asc" | "desc" }[];
}

export const organizationRolesApi = {
	list: (orgId: string) => {
		return api.get<{ data: Role[] }>(`/organizations/${orgId}/roles`);
	},
	create: (orgId: string, data: { name: string; description: string }) => {
		return api.post<{ data: Role }>(`/organizations/${orgId}/roles`, data);
	},
	update: (orgId: string, roleId: string, data: { description: string }) => {
		return api.put<{ data: Role }>(
			`/organizations/${orgId}/roles/${roleId}`,
			data,
		);
	},
	delete: (orgId: string, roleId: string) => {
		return api.delete(`/organizations/${orgId}/roles/${roleId}`);
	},
};

export const rolesApi = {
	// Get all roles
	getAll: () => {
		return api.get<RoleListResponse>("/roles");
	},

	// Create role
	create: (data: { name: string; description: string }) => {
		return api.post<{ data: Role }>("/roles", data);
	},

	// Update role
	update: (id: string, data: { description: string }) => {
		return api.put<{ data: Role }>(`/roles/${id}`, data);
	},

	// Delete role
	delete: (id: string) => {
		return api.delete(`/roles/${id}`);
	},

	// Search roles
	search: (filter: SearchRoleFilter) => {
		return api.post<RoleListResponse>("/roles/search", filter);
	},
};
