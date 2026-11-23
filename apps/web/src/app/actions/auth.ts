"use server";

import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { redirect } from "next/navigation";
import { loginSchema } from "~/lib/api/auth";
import { actionClient } from "~/lib/client/safe-action";

const BACKEND_URL =
	process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080/api/v1";

const isSecure =
	process.env.NEXT_PUBLIC_COOKIE_SECURE === "true" ||
	(process.env.NODE_ENV === "production" &&
		process.env.NEXT_PUBLIC_COOKIE_SECURE !== "false");

const ACCESS_TOKEN_MAX_AGE = Number(
	process.env.NEXT_PUBLIC_ACCESS_TOKEN_MAX_AGE || 60 * 15,
);
const REFRESH_TOKEN_MAX_AGE = Number(
	process.env.NEXT_PUBLIC_REFRESH_TOKEN_MAX_AGE || 60 * 60 * 24,
);

export const loginAction = actionClient
	.schema(loginSchema)
	.metadata({ actionName: "login" })
	.action(async ({ parsedInput: { username, password } }) => {
		try {
			let response: Response;
			try {
				response = await fetch(`${BACKEND_URL}/auth/login`, {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({ username, password }),
				});
			} catch (fetchError) {
				console.error("Login Fetch Error:", fetchError);
				return {
					success: false,
					message:
						"Sistem tidak dapat terhubung ke API server. Pastikan backend sudah menyala.",
				};
			}

			const result = await response.json();

			if (!response.ok) {
				if (response.status === 502 || result.code === "BACKEND_OFFLINE") {
					throw new Error(
						result.message || "Gagal terhubung ke server (Backend Offline).",
					);
				}
				throw new Error(result.error || result.message || "Login failed");
			}

			const { data } = result;
			const cookieStore = await cookies();
			const upstream = new NextResponse(null, { headers: response.headers })
				.cookies;
			const csrf = upstream.get("csrf_token");
			const refresh = upstream.get("refresh_token");
			if (!csrf || !refresh) throw new Error("Missing session cookies");
			cookieStore.set("csrf_token", csrf.value, {
				secure: isSecure,
				sameSite: "lax",
				path: "/",
				maxAge: REFRESH_TOKEN_MAX_AGE,
			});

			cookieStore.set("access_token", data.access_token, {
				httpOnly: true,
				secure: isSecure,
				sameSite: "lax",
				path: "/",
				maxAge: ACCESS_TOKEN_MAX_AGE,
			});

			cookieStore.set("refresh_token", refresh.value, {
				httpOnly: true,
				secure: isSecure,
				sameSite: "lax",
				path: "/",
				maxAge: REFRESH_TOKEN_MAX_AGE,
			});

			return { success: true, user: data.user };
		} catch (error: any) {
			return { success: false, message: error.message };
		}
	});

export const logoutAction = async () => {
	const cookieStore = await cookies();
	cookieStore.delete("access_token");
	cookieStore.delete("refresh_token");
	cookieStore.delete("csrf_token");
	redirect("/login");
};
