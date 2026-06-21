import { apiClient } from "./client";
import {
  userSchema,
  paginatedSchema,
  type User,
  type PaginatedResponse,
} from "./schemas";

const paginatedUsers = paginatedSchema(userSchema);

export const usersApi = {
  list: (params?: { page?: number; limit?: number; search?: string }) =>
    apiClient.get<PaginatedResponse<User>>("/users", paginatedUsers, {
      params,
    }),

  get: (id: string) => apiClient.get<User>(`/users/${id}`, userSchema),

  delete: (id: string) => apiClient.delete(`/users/${id}`),
};
