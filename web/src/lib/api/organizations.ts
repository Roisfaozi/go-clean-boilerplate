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

  // --- Member Management ---

  getMembers: (orgId: string) => {
    return api.get<{ data: Member[] }>(`/organizations/${orgId}/members`);
  },

  inviteMember: (orgId: string, data: { email: string; role_id: string }) => {
    return api.post<{ data: Member }>(`/organizations/${orgId}/members/invite`, data);
  },

  updateMemberRole: (orgId: string, userId: string, data: { role_id: string }) => {
    return api.patch<{ data: Member }>(`/organizations/${orgId}/members/${userId}`, data);
  },

  removeMember: (orgId: string, userId: string) => {
    return api.delete(`/organizations/${orgId}/members/${userId}`);
  },

  acceptInvitation: (data: { token: string; password?: string; name?: string }) => {
    return api.post("/organizations/invitations/accept", data);
  },

  getPresence: (orgId: string) => {
    return api.get<{ data: Member[] }>(`/organizations/${orgId}/presence`);
  },
};

export interface Member {
  id: string;
  organization_id: string;
  user_id: string;
  user?: {
    id: string;
    name: string;
    email: string;
    avatar_url?: string;
  };
  role_id: string;
  role?: {
    id: string;
    name: string;
  };
  status: string;
  joined_at: number;
}
