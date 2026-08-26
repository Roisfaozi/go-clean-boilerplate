import { apiClient } from "./client";
import {
	loginRequestSchema,
	loginResponseSchema,
	registerRequestSchema,
	tokenResponseSchema,
	userSchema,
	type LoginRequest,
	type LoginResponse,
	type RegisterRequest,
	type TokenResponse,
	type User,
} from "./schemas";

export interface WsTicketResponse {
	ticket: string;
	expires_in: number;
}

export const authApi = {
	login: (data: LoginRequest) => {
		loginRequestSchema.parse(data);
		return apiClient.post<LoginResponse>(
			"/auth/login",
			data,
			loginResponseSchema,
		);
	},

	register: (data: RegisterRequest) => {
		registerRequestSchema.parse(data);
		return apiClient.post<LoginResponse>(
			"/auth/register",
			data,
			loginResponseSchema,
		);
	},

	logout: () => apiClient.post("/auth/logout"),

	refreshToken: () =>
		apiClient.post<TokenResponse>(
			"/auth/refresh",
			undefined,
			tokenResponseSchema,
		),

	getWsTicket: () => apiClient.post<WsTicketResponse>("/auth/ticket"),

	getCurrentUser: () => apiClient.get<User>("/auth/me", userSchema),

	forgotPassword: (email: string) =>
		apiClient.post("/auth/forgot-password", { email }),

	resetPassword: (token: string, newPassword: string) =>
		apiClient.post("/auth/reset-password", {
			token,
			new_password: newPassword,
		}),
};
