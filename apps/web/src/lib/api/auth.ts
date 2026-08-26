import { z } from "zod";
import { api } from "./client";

export const loginSchema = z.object({
	username: z.string().min(3, "Username must be at least 3 characters."),
	password: z.string().min(1, "Password is required."),
});

export type LoginInput = z.infer<typeof loginSchema>;

export const forgotPasswordSchema = z.object({
	email: z.string().email("Please enter a valid email address.").max(100),
});

export const authTokenSchema = z.string().min(1).max(500);

export const resetPasswordSchema = z
	.object({
		password: z
			.string()
			.min(8, "Password must be at least 8 characters.")
			.max(72, "Password must be at most 72 characters."),
		confirmPassword: z.string(),
	})
	.refine((data) => data.password === data.confirmPassword, {
		message: "Passwords do not match.",
		path: ["confirmPassword"],
	});

export type ForgotPasswordInput = z.infer<typeof forgotPasswordSchema>;
export type ResetPasswordInput = z.infer<typeof resetPasswordSchema>;

interface MessageResponse {
	data: {
		message: string;
	};
}

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

	forgotPassword: (data: ForgotPasswordInput) => {
		return api.post<MessageResponse>("/auth/forgot-password", data);
	},

	resetPassword: (token: string, newPassword: string) => {
		return api.post<MessageResponse>("/auth/reset-password", {
			token,
			new_password: newPassword,
		});
	},

	verifyEmail: (token: string) => {
		return api.post<MessageResponse>("/auth/verify-email", { token });
	},

	/**
	 * Cek user yang sedang login.
	 * Menggunakan silentGet agar 401 tidak memicu redirect ke /login
	 * (penting untuk public pages seperti landing page).
	 */
	getCurrentUser: () => {
		return api.silentGet<{ data: { user: any } }>("/auth/me");
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
