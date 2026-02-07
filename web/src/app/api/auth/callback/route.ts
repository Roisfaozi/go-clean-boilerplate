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

  // Save Access Token (Short-lived)
  cookieStore.set("access_token", token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 60 * 15, // 15 minutes
  });

  // Save Refresh Token (Long-lived)
  if (refreshToken) {
    cookieStore.set("refresh_token", refreshToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: 60 * 60 * 24 * 30, // 30 days
    });
  }

  return NextResponse.redirect(
    new URL(decodeURIComponent(returnTo), request.url)
  );
}
