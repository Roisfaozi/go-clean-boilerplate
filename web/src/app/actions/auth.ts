"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { actionClient } from "~/lib/client/safe-action";
import { loginSchema } from "~/lib/api/auth";

const BACKEND_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export const loginAction = actionClient
  .metadata({ actionName: "login" })
  .schema(loginSchema)
  .action(async ({ parsedInput: { username, password } }) => {
    try {
      const response = await fetch(`${BACKEND_URL}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || result.message || "Login failed");
      }

      const { data } = result;
      const cookieStore = await cookies();

      // Set Access Token
      cookieStore.set("access_token", data.access_token, {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        path: "/",
        maxAge: 60 * 15,
      });

      // Set Refresh Token (extract from Set-Cookie header of backend or use JSON if provided)
      // Note: In our previous check, backend sets refresh_token cookie.
      // But since Next.js is proxying, we might need to manually set it if fetch doesn't forward it.
      // For now, let's assume it's in the data response as per authApi.ts interface
      if (data.refresh_token) {
        cookieStore.set("refresh_token", data.refresh_token, {
          httpOnly: true,
          secure: process.env.NODE_ENV === "production",
          sameSite: "lax",
          path: "/",
          maxAge: 60 * 60 * 24 * 30,
        });
      }

      return { success: true, user: data.user };
    } catch (error: any) {
      return { success: false, message: error.message };
    }
  });

export const logoutAction = async () => {
  const cookieStore = await cookies();
  cookieStore.delete("access_token");
  cookieStore.delete("refresh_token");
  redirect("/login");
};
