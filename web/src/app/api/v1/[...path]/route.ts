import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

const BACKEND_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080/api/v1";

export async function ALL(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const resolvedParams = await params;
  const path = resolvedParams.path.join("/");
  const url = `${BACKEND_URL}/${path}${request.nextUrl.search}`;
  const cookieStore = await cookies();
  const token = cookieStore.get("access_token")?.value;

  const headers = new Headers(request.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  // Explicitly forward all cookies from Next.js cookie store to the backend
  const allCookies = cookieStore.getAll();
  if (allCookies.length > 0) {
    const cookieHeader = allCookies
      .map((c) => `${c.name}=${c.value}`)
      .join("; ");
    headers.set("Cookie", cookieHeader);
  }

  // Remove host header to avoid conflicts with backend
  headers.delete("host");

  try {
    const body =
      request.method !== "GET" && request.method !== "HEAD"
        ? await request.blob()
        : undefined;

    const response = await fetch(url, {
      method: request.method,
      headers: headers,
      body: body,
      cache: "no-store",
    });

    // Create response headers
    const responseHeaders = new Headers();
    response.headers.forEach((value, key) => {
      // Forward Set-Cookie and other safe headers
      if (
        key.toLowerCase() === "set-cookie" ||
        key.toLowerCase() === "content-type" ||
        key.toLowerCase() === "cache-control"
      ) {
        responseHeaders.append(key, value);
      }
    });

    const contentType = response.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      const data = await response.json();
      return NextResponse.json(data, {
        status: response.status,
        headers: responseHeaders,
      });
    }

    const data = await response.text();
    return new NextResponse(data, {
      status: response.status,
      headers: responseHeaders,
    });
  } catch (error) {
    console.error("Proxy Error (Backend Unreachable):", error);
    return NextResponse.json(
      {
        success: false,
        code: "BACKEND_OFFLINE",
        message:
          "Gagal terhubung ke API Server. Pastikan backend sudah menyala di port 8080.",
      },
      { status: 502 }
    );
  }
}

export { ALL as DELETE, ALL as GET, ALL as PATCH, ALL as POST, ALL as PUT };
