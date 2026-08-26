import { apiClient } from "@/lib/api/client";
import type { PaginatedResponse } from "@/lib/api/schemas";

export interface AuditLog {
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

export const auditService = {
	search: (params: {
		page?: number;
		limit?: number;
		sort?: string;
		filters?: any[];
		search?: string;
	}) =>
		apiClient.post<PaginatedResponse<AuditLog>>("/audit-logs/search", params),

	export: (params: {
		from_date?: string;
		to_date?: string;
		format: "csv" | "excel";
	}) =>
		apiClient.get("/audit-logs/export", undefined, {
			params,
			responseType: "blob",
		}),
};
