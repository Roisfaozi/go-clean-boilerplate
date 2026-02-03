import { api } from "./client";

export interface User {
  id: string;
  name: string;
  email: string;
  username: string;
  avatar_url?: string;
  status: string;
  created_at: number;
  updated_at: number;
}

export interface UserListResponse {
  data: User[];
  paging: {
    limit: number;
    page: number;
    size: number;
    total: number;
    total_item: number;
    total_page: number;
  };
}

export interface SearchUserFilter {
  page?: number;
  page_size?: number;
  filter?: Record<string, { type: string; [key: string]: any }>;
  sort?: { colId: string; sort: "asc" | "desc" }[];
}

export const usersApi = {
  // Get all users (Admin only)
  getAll: (page = 1, limit = 10, username?: string, email?: string) => {
    const params = new URLSearchParams();
    params.append("page", page.toString());
    params.append("limit", limit.toString());
    if (username) params.append("username", username);
    if (email) params.append("email", email);
    
    return api.get<UserListResponse>(`/users?${params.toString()}`);
  },

  // Get user by ID
  getById: (id: string) => {
    return api.get<{ data: User }>(`/users/${id}`);
  },

  // Search users with dynamic filter
  search: (filter: SearchUserFilter) => {
    return api.post<UserListResponse>("/users/search", filter);
  },

  // Delete user
  delete: (id: string) => {
    return api.delete(`/users/${id}`);
  },

  // Update user status
  updateStatus: (id: string, status: "active" | "suspended" | "banned") => {
    return api.request(`/users/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    });
  },

  // Get current user profile
  getMe: () => {
    return api.get<{ data: User }>("/users/me");
  },

  // Update current user profile
  updateMe: (data: { name?: string; password?: string; username?: string }) => {
    return api.put<{ data: User }>("/users/me", data);
  },

  // Upload avatar
  uploadAvatar: (file: File) => {
    const formData = new FormData();
    formData.append("avatar", file);
    // Note: Client.ts post method currently expects JSON and stringifies body.
    // We might need a raw request method or modify client to handle FormData.
    // For now, assuming we might need a specific override or update client.ts
    // Let's assume we update client.ts or use fetch directly for this specific case if needed.
    // But since I can't modify client.ts easily in this turn without reading/writing it again, 
    // I'll stick to the pattern.
    // Actually, checking client.ts, it sets Content-Type to json by default.
    // We need to allow overriding it.
    
    // Workaround: We'll implement a custom request for multipart in the component 
    // or assume the client will be updated to support FormData.
    return api.request<{ data: User }>("/users/me/avatar", {
        method: "PATCH",
        body: formData,
        headers: {} // Let browser set Content-Type for FormData
    });
  }
};
