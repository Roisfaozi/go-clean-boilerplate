import { apiClient } from "./client";
import {
	organizationSchema,
	orgMemberSchema,
	paginatedSchema,
	type Organization,
	type OrgMember,
	type PaginatedResponse,
} from "./schemas";

const paginatedOrgs = paginatedSchema(organizationSchema);
const paginatedMembers = paginatedSchema(orgMemberSchema);

export const organizationsApi = {
	list: (params?: { page?: number; limit?: number }) =>
		apiClient.get<PaginatedResponse<Organization>>(
			"/organizations/me",
			paginatedOrgs,
			{ params },
		),

	create: (data: { name: string; slug?: string }) =>
		apiClient.post<Organization>("/organizations", data, organizationSchema),

	update: (id: string, data: Partial<Organization>) =>
		apiClient.put<Organization>(
			`/organizations/${id}`,
			data,
			organizationSchema,
		),

	delete: (id: string) => apiClient.delete(`/organizations/${id}`),

	getMembers: (orgId: string) =>
		apiClient.get<PaginatedResponse<OrgMember>>(
			`/organizations/${orgId}/members`,
			paginatedMembers,
		),

	inviteMember: (orgId: string, data: { email: string; role_id: string }) =>
		apiClient.post(`/organizations/${orgId}/members/invite`, data),

	removeMember: (orgId: string, memberId: string) =>
		apiClient.delete(`/organizations/${orgId}/members/${memberId}`),
};
