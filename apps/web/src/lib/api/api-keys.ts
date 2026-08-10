import { api } from "./client";

export interface ApiKey {
	id: string;
	name: string;
	organization_id: string;
	user_id: string;
	scopes: string[];
	expires_at: string | null;
	last_used_at: string | null;
	is_active: boolean;
	created_at: string;
}

export interface ApiKeyCreated extends ApiKey {
	api_key: string;
}

export interface CreateApiKeyRequest {
	name: string;
	scopes: string[];
	expires_at?: string | null;
}

export interface ApiKeyListResponse {
	data: ApiKey[];
}

export const apiKeysApi = {
	list: () => api.get<ApiKeyListResponse>("/api-keys").then((res) => res.data),

	create: (req: CreateApiKeyRequest) =>
		api.post<{ data: ApiKeyCreated }>("/api-keys", req).then((res) => res.data),

	revoke: (id: string) => api.delete(`/api-keys/${id}`),
};
