import { cookies } from "next/headers";
import { NextResponse, type NextRequest } from "next/server";

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const token = searchParams.get("token");
  const refreshToken = searchParams.get("refresh_token");
  const returnTo = searchParams.get("returnTo") || "/dashboard";

  if (!token) {
    return NextResponse.redirect(
      new URL("/login?error=unauthorized", request.url)
    );
  }

  const cookieStore = await cookies();

  const isSecure =
    process.env.NEXT_PUBLIC_COOKIE_SECURE === "true" ||
    (process.env.NODE_ENV === "production" &&
      process.env.NEXT_PUBLIC_COOKIE_SECURE !== "false");

  const ACCESS_TOKEN_MAX_AGE = Number(
    process.env.NEXT_PUBLIC_ACCESS_TOKEN_MAX_AGE || 60 * 15
  );
  const REFRESH_TOKEN_MAX_AGE = Number(
    process.env.NEXT_PUBLIC_REFRESH_TOKEN_MAX_AGE || 60 * 60 * 24
  );

  cookieStore.set("access_token", token, {
    httpOnly: true,
    secure: isSecure,
    sameSite: "lax",
    path: "/",
    maxAge: ACCESS_TOKEN_MAX_AGE,
  });

  if (refreshToken) {
    cookieStore.set("refresh_token", refreshToken, {
      httpOnly: true,
      secure: isSecure,
      sameSite: "lax",
      path: "/",
      maxAge: REFRESH_TOKEN_MAX_AGE,
    });
  }

  return NextResponse.redirect(
    new URL(decodeURIComponent(returnTo), request.url)
  );
}
