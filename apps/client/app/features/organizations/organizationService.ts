import { apiClient } from "@/lib/api/client";
import type { Organization } from "@/lib/api/schemas";

export const organizationService = {
	list: (params?: { page?: number; limit?: number }) =>
		apiClient.get<{ organizations: Organization[]; total: number }>(
			"/organizations/me",
			undefined,
			{ params },
		),
	create: (data: { name: string; slug?: string }) =>
		apiClient.post<Organization>("/organizations", data),
	update: (id: string, data: Pick<Organization, "name" | "status">) =>
		apiClient.put<Organization>(`/organizations/${id}`, data, undefined, {
			headers: { "X-Organization-ID": id },
		}),
	delete: (id: string) =>
		apiClient.delete(`/organizations/${id}`, {
			headers: { "X-Organization-ID": id },
		}),
};
