import { useAuthStore } from "@/stores/auth-store";
import { useOrganizationStore } from "@/stores/organization-store";
import axios, { type AxiosError, type AxiosRequestConfig } from "axios";
import type { z } from "zod";

const API_BASE_URL = "/api/v1";

// ── API Error ──
export class ApiError extends Error {
	constructor(
		message: string,
		public status: number,
		public code?: string,
		public details?: unknown,
	) {
		super(message);
		this.name = "ApiError";
	}

	static fromAxios(
		error: AxiosError<{ message?: string; error?: string; code?: string }>,
	): ApiError {
		const status = error.response?.status ?? 0;
		const message =
			error.response?.data?.message ??
			error.response?.data?.error ??
			error.message;
		const code = error.response?.data?.code;
		return new ApiError(message, status, code, error.response?.data);
	}

	get isUnauthorized() {
		return this.status === 401;
	}
	get isForbidden() {
		return this.status === 403;
	}
	get isNotFound() {
		return this.status === 404;
	}
	get isValidation() {
		return this.status === 422;
	}
	get isServer() {
		return this.status >= 500;
	}
}

// ── Axios instance ──
const axiosInstance = axios.create({
	baseURL: API_BASE_URL,
	withCredentials: true,
	headers: {
		"Content-Type": "application/json",
		Accept: "application/json",
	},
});

axiosInstance.interceptors.request.use((config) => {
	const activeOrganization = useOrganizationStore.getState().activeOrganization;
	if (activeOrganization && config.headers) {
		if (!config.headers["X-Organization-ID"]) {
			config.headers["X-Organization-ID"] = activeOrganization.id;
		}
		if (!config.headers["X-Organization-Slug"]) {
			config.headers["X-Organization-Slug"] = activeOrganization.slug;
		}
	}

	return config;
});

let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

axiosInstance.interceptors.response.use(
	(response) => response,
	async (error: AxiosError) => {
		const originalRequest = error.config as AxiosRequestConfig & {
			_retry?: boolean;
		};
		const status = error.response?.status;

		if (
			status === 401 &&
			originalRequest &&
			!originalRequest._retry &&
			!originalRequest.url?.includes("/auth/refresh") &&
			!originalRequest.url?.includes("/auth/login")
		) {
			if (!isRefreshing) {
				isRefreshing = true;
				refreshPromise = (async () => {
					try {
						await axiosInstance.post("/auth/refresh");
						return true;
					} catch {
						return false;
					} finally {
						isRefreshing = false;
						refreshPromise = null;
					}
				})();
			}

			const refreshed = await refreshPromise;
			if (refreshed && originalRequest) {
				originalRequest._retry = true;
				return axiosInstance(originalRequest);
			}

			if (typeof window !== "undefined") {
				const isLoginPage = window.location.pathname === "/login";
				if (!isLoginPage) {
					useAuthStore.getState().logout();
					window.location.href = "/login";
				}
			}
		}

		return Promise.reject(
			ApiError.fromAxios(
				error as AxiosError<{
					message?: string;
					error?: string;
					code?: string;
				}>,
			),
		);
	},
);

// ── Type-safe request helpers ─

function extractPayload<T>(res: { data: unknown }): T {
	const raw = res.data as Record<string, unknown>;
	if (raw && typeof raw === "object" && "data" in raw) {
		return raw.data as T;
	}
	return res.data as T;
}

/** Validates response data against a Zod schema in development, passes through in production. */
function validate<T>(schema: z.ZodType<T> | z.ZodTypeAny, data: unknown): T {
	if (import.meta.env.DEV) {
		const result = schema.safeParse(data);
		if (!result.success) {
			console.warn(
				"[API] Response validation warning:",
				result.error.flatten(),
			);
		}
	}
	return data as T;
}

export const apiClient = {
	async get<T>(
		url: string,
		schema?: z.ZodTypeAny,
		config?: AxiosRequestConfig,
	): Promise<T> {
		const res = await axiosInstance.get(url, config);
		const payload = extractPayload<T>(res);
		return schema ? validate<T>(schema, payload) : payload;
	},

	async post<T>(
		url: string,
		data?: unknown,
		schema?: z.ZodTypeAny,
		config?: AxiosRequestConfig,
	): Promise<T> {
		const res = await axiosInstance.post(url, data, config);
		const payload = extractPayload<T>(res);
		return schema ? validate<T>(schema, payload) : payload;
	},

	async put<T>(
		url: string,
		data?: unknown,
		schema?: z.ZodTypeAny,
		config?: AxiosRequestConfig,
	): Promise<T> {
		const res = await axiosInstance.put(url, data, config);
		const payload = extractPayload<T>(res);
		return schema ? validate<T>(schema, payload) : payload;
	},

	async patch<T>(
		url: string,
		data?: unknown,
		schema?: z.ZodTypeAny,
		config?: AxiosRequestConfig,
	): Promise<T> {
		const res = await axiosInstance.patch(url, data, config);
		const payload = extractPayload<T>(res);
		return schema ? validate<T>(schema, payload) : payload;
	},

	async delete<T = void>(url: string, config?: AxiosRequestConfig): Promise<T> {
		const res = await axiosInstance.delete(url, config);
		return extractPayload<T>(res);
	},
};

export default apiClient;
