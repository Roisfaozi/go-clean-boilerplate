import { api } from "./client";

export interface Webhook {
	id: string;
	name: string;
	organization_id: string;
	url: string;
	events: string[];
	is_active: boolean;
	created_at: number;
	updated_at: number;
}

export interface WebhookLog {
	id: string;
	webhook_id: string;
	event_type: string;
	payload: string;
	response_status_code: number;
	response_body: string;
	execution_time: number;
	error_message: string;
	retry_count: number;
	created_at: number;
}

export interface CreateWebhookRequest {
	name: string;
	url: string;
	events: string[];
	secret: string;
}

export interface UpdateWebhookRequest {
	name?: string;
	url?: string;
	events?: string[];
	secret?: string;
	is_active?: boolean;
}

export interface WebhookListResponse {
	data: Webhook[];
}

export interface WebhookLogListResponse {
	data: WebhookLog[];
}

export const webhooksApi = {
	list: () => api.get<WebhookListResponse>("/webhooks").then((res) => res.data),

	getByID: (id: string) =>
		api.get<{ data: Webhook }>(`/webhooks/${id}`).then((res) => res.data),

	create: (req: CreateWebhookRequest) =>
		api.post<{ data: Webhook }>("/webhooks", req).then((res) => res.data),

	update: (id: string, req: UpdateWebhookRequest) =>
		api.put<{ data: Webhook }>(`/webhooks/${id}`, req).then((res) => res.data),

	delete: (id: string) => api.delete(`/webhooks/${id}`),

	logs: (id: string) =>
		api
			.get<WebhookLogListResponse>(`/webhooks/${id}/logs`)
			.then((res) => res.data),
};
