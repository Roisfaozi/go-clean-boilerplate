import { splitCookiesString } from "next/dist/server/web/utils";
import { type NextRequest, NextResponse } from "next/server";

const BACKEND_URL =
	process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080/api/v1";

async function ALL(
	request: NextRequest,
	{ params }: { params: Promise<{ path: string[] }> },
) {
	const resolvedParams = await params;
	const path = resolvedParams.path.join("/");
	const url = `${BACKEND_URL}/${path}${request.nextUrl.search}`;
	const headers = new Headers(request.headers);

	// Remove host header to avoid conflicts with backend
	headers.delete("host");

	try {
		const body =
			request.method !== "GET" && request.method !== "HEAD"
				? await request.blob()
				: undefined;

		const response = await fetch(url, {
			method: request.method,
			headers,
			body,
			cache: "no-store",
			signal: request.signal,
			redirect: "manual",
		});

		const responseHeaders = new Headers();
		response.headers.forEach((value, key) => {
			if (
				[
					"content-type",
					"cache-control",
					"content-disposition",
					"x-request-id",
					"location",
					"retry-after",
				].includes(key) ||
				key.startsWith("x-ratelimit-")
			) {
				responseHeaders.append(key, value);
			}
		});

		const cookies =
			response.headers.getSetCookie?.() ??
			splitCookiesString(response.headers.get("set-cookie") ?? "");
		for (const cookie of cookies) {
			responseHeaders.append("set-cookie", cookie);
		}

		return new Response(response.body, {
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
			{ status: 502 },
		);
	}
}

export { ALL as DELETE, ALL as GET, ALL as PATCH, ALL as POST, ALL as PUT };
