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
  const token = (await cookies()).get("access_token")?.value;

  const headers = new Headers(request.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
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

    const contentType = response.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      const data = await response.json();
      return NextResponse.json(data, { status: response.status });
    }

    const data = await response.text();
    return new NextResponse(data, {
      status: response.status,
      headers: {
        "Content-Type": contentType || "text/plain",
      },
    });
  } catch (error) {
    console.error("Proxy Error:", error);
    return NextResponse.json(
      { error: "Backend server unreachable" },
      { status: 502 }
    );
  }
}

export { ALL as DELETE, ALL as GET, ALL as PATCH, ALL as POST, ALL as PUT };
