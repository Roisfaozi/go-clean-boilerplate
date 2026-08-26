import { apiClient } from "./client";
import type { PaginatedResponse } from "./types";

export interface AuditLogResponse {
	id: string;
	organization_id: string | null;
	user_id: string;
	action: string;
	entity: string;
	entity_id: string;
	ip_address: string;
	user_agent: string;
	created_at: number;
}

export const auditApi = {
	search: (params: {
		page?: number;
		limit?: number;
		sort?: string;
		filters?: any[];
	}) =>
		apiClient.post<PaginatedResponse<AuditLogResponse>>(
			"/audit-logs/search",
			params,
		),

	export: (params: {
		from_date?: string;
		to_date?: string;
		format: "csv" | "excel";
	}) => apiClient.get("/audit-logs/export", undefined, { params }),
};
