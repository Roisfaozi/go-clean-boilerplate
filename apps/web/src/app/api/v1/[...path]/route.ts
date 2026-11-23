import { type NextRequest, NextResponse } from "next/server";

const BACKEND_URL =
	process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080/api/v1";

function splitSetCookieHeader(header: string): string[] {
	if (!header) return [];
	const cookies: string[] = [];
	let start = 0;
	let pos = 0;
	while (pos < header.length) {
		const commaIdx = header.indexOf(",", pos);
		if (commaIdx === -1) break;
		const afterComma = header.slice(commaIdx + 1).trimStart();
		if (/^[^\s;=,]+=/.test(afterComma)) {
			cookies.push(header.slice(start, commaIdx).trim());
			start = header.length - afterComma.length;
		}
		pos = commaIdx + 1;
	}
	const tail = header.slice(start).trim();
	if (tail) cookies.push(tail);
	return cookies;
}

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
			splitSetCookieHeader(response.headers.get("set-cookie") ?? "");
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

export {
	ALL as DELETE,
	ALL as GET,
	ALL as HEAD,
	ALL as PATCH,
	ALL as POST,
	ALL as PUT,
};
