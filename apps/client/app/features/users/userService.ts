import { apiClient } from "@/lib/api/client";
import {
	userSchema,
	paginatedSchema,
	type User,
	type PaginatedResponse,
} from "@/lib/api/schemas";

const paginatedUsers = paginatedSchema(userSchema);

export const userService = {
	list: (params?: { page?: number; limit?: number; search?: string }) =>
		apiClient.get<PaginatedResponse<User>>("/users", paginatedUsers, {
			params,
		}),

	get: (id: string) => apiClient.get<User>(`/users/${id}`, userSchema),

	create: (data: {
		name: string;
		email: string;
		username: string;
		password: string;
		role?: string;
	}) => apiClient.post<User>("/users", data, userSchema),

	update: (id: string, data: Partial<User>) =>
		apiClient.put<User>(`/users/${id}`, data, userSchema),

	delete: (id: string) => apiClient.delete(`/users/${id}`),
};
