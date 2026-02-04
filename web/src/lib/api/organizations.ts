import { api } from "./client";

export interface Organization {
  id: string;
  name: string;
  slug: string;
  status: string;
  owner_id: string;
  created_at: number;
  updated_at: number;
}

export const organizationsApi = {
  create: (data: { name: string; slug: string }) => {
    return api.post<{ data: Organization }>("/organizations", data);
  },

  getMyOrganizations: () => {
    return api.get<{ data: { organizations: Organization[]; total: number } }>(
      "/organizations/me"
    );
  },

  getBySlug: (slug: string) => {
    return api.get<{ data: Organization }>(`/organizations/slug/${slug}`);
  },

  getById: (id: string) => {
    return api.get<{ data: Organization }>(`/organizations/${id}`);
  },

  update: (
    id: string,
    data: { name?: string; status?: "active" | "suspended" | "inactive" }
  ) => {
    return api.put<{ data: Organization }>(`/organizations/${id}`, data);
  },

  delete: (id: string) => {
    return api.delete(`/organizations/${id}`);
  },
};
