import { cookies } from "next/headers";

const isSecure =
  process.env.NEXT_PUBLIC_COOKIE_SECURE === "true" ||
  (process.env.NODE_ENV === "production" &&
    process.env.NEXT_PUBLIC_COOKIE_SECURE !== "false");

export const setSessionTokenCookie = async (token: string, expiresAt: Date) => {
  const cookieStore = await cookies();
  cookieStore.set("session", token, {
    httpOnly: true,
    sameSite: "lax",
    secure: isSecure,
    expires: expiresAt,
    path: "/",
  });
};

export const deleteSessionTokenCookie = async () => {
  const cookieStore = await cookies();
  cookieStore.set("session", "", {
    httpOnly: true,
    sameSite: "lax",
    secure: isSecure,
    maxAge: 0,
    path: "/",
  });
};
