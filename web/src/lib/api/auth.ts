import { z } from "zod";
import { api } from "./client";

export const loginSchema = z.object({
  username: z.string().min(3, "Username must be at least 3 characters."),
  password: z.string().min(1, "Password is required."),
});

export type LoginInput = z.infer<typeof loginSchema>;

export interface AuthResponse {
  data: {
    access_token: string;
    token_type: string;
    expires_in: number;
    refresh_token: string;
    expires_at: string;
    user: {
      id: string;
      name: string;
      email: string;
      username: string;
      role: string;
    };
  };
}

export const authApi = {
  login: (data: LoginInput) => {
    return api.post<AuthResponse>("/auth/login", data);
  },

  logout: () => {
    return api.post("/auth/logout", {});
  },

  register: (data: any) => {
    return api.post<AuthResponse>("/auth/register", data);
  },

  /**
   * Cek user yang sedang login.
   * Menggunakan silentGet agar 401 tidak memicu redirect ke /login
   * (penting untuk public pages seperti landing page).
   */
  getCurrentUser: () => {
    return api.silentGet<{ user: any }>("/auth/me");
  },

  resendVerification: () => {
    return api.post("/auth/resend-verification", {});
  },

  getWsTicket: (orgId?: string) => {
    const url = orgId ? `/auth/ticket?org_id=${orgId}` : "/auth/ticket";
    return api
      .post<{ data: { ticket: string } }>(url, {})
      .then((res) => res.data);
  },
};
