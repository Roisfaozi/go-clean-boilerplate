import { apiClient } from "./client";

export interface PermissionResponse {
	role: string;
	domain: string;
	path: string;
	method: string;
}

export interface RoleNode {
	id: string;
	name: string;
	description?: string;
	parent_id?: string | null;
	parents?: string[];
	children?: RoleNode[];
	own_permissions?: string[][];
	inherited_permissions?: string[][];
	effective_permissions?: string[][];
}

export interface InheritanceTreeResponse {
	roles: RoleNode[];
}

export interface ResourceCRUD {
	create: boolean;
	read: boolean;
	update: boolean;
	delete: boolean;
}

export interface ResourcePermission {
	name: string;
	base_path: string;
	role_permissions: Record<string, ResourceCRUD>;
	endpoint_count: number;
}

export interface ResourceAggregationResponse {
	resources: ResourcePermission[];
}

export const permissionsApi = {
	getInheritanceTree: () =>
		apiClient.get<InheritanceTreeResponse>("/permissions/inheritance-tree"),

	getResourceAggregation: () =>
		apiClient.get<ResourceAggregationResponse>("/permissions/resources"),

	grant: (data: {
		role: string;
		path: string;
		method: string;
		domain?: string;
	}) => apiClient.post("/permissions/grant", data),

	revoke: (data: {
		role: string;
		path: string;
		method: string;
		domain?: string;
	}) => apiClient.post("/permissions/revoke", data),

	addInheritance: (data: { role: string; parent: string; domain?: string }) =>
		apiClient.post("/permissions/inheritance", data),

	removeInheritance: (data: {
		role: string;
		parent: string;
		domain?: string;
	}) => apiClient.delete("/permissions/inheritance", { data }),
};
