// @vitest-environment node
import { NextRequest } from "next/server";
import { afterEach, expect, it, vi } from "vitest";
import { GET, HEAD, POST } from "./route";

const params = Promise.resolve({ path: ["events"] });
const url = "http://localhost/api/v1/events";
afterEach(() => vi.unstubAllGlobals());

it.each([
	true,
	false,
])("preserves separate cookies and request headers (native cookie API %s)", async (native) => {
	const cookies = [
		"access_token=a; HttpOnly",
		"refresh_token=b; Expires=Wed, 21 Oct 2037 07:28:00 GMT",
		"csrf_token=c; Path=/",
	];
	const upstream = new Response(null, { status: 204 });
	for (const cookie of cookies) upstream.headers.append("set-cookie", cookie);
	if (!native)
		Object.defineProperty(upstream.headers, "getSetCookie", {
			value: undefined,
		});
	const fetch = vi.fn<typeof globalThis.fetch>().mockResolvedValue(upstream);
	vi.stubGlobal("fetch", fetch);
	const headers = {
		Cookie: "access_token=a; csrf_token=c",
		"X-CSRF-Token": "c",
		"X-Request-ID": "trace",
		"Idempotency-Key": "key",
		"X-Organization-ID": "org",
		"X-Organization-Slug": "slug",
	};
	const result = await POST(new NextRequest(url, { method: "POST", headers }), {
		params,
	});
	const forwarded = new Headers(fetch.mock.calls[0][1]?.headers);
	for (const [key, value] of Object.entries(headers))
		expect(forwarded.get(key)).toBe(value);
	expect(forwarded.has("authorization")).toBe(false);
	expect(result.headers.getSetCookie()).toEqual(cookies);
	expect(result.status).toBe(204);
	expect(result.body).toBeNull();
});

it("returns the open SSE stream before it closes", async () => {
	const stream = new ReadableStream<Uint8Array>({
		start(controller) {
			controller.enqueue(new TextEncoder().encode("data: live\n\n"));
		},
	});
	const headers = {
		"content-type": "text/event-stream",
		"cache-control": "no-cache",
		"x-request-id": "trace",
		"x-ratelimit-remaining": "7",
	};
	vi.stubGlobal(
		"fetch",
		vi.fn().mockResolvedValue(new Response(stream, { headers })),
	);
	const result = await GET(new NextRequest(url), { params });
	expect(result.body).toBe(stream);
	for (const [key, value] of Object.entries(headers))
		expect(result.headers.get(key)).toBe(value);
	const reader = result.body!.getReader();
	expect(new TextDecoder().decode((await reader.read()).value)).toBe(
		"data: live\n\n",
	);
	await reader.cancel();
});

it("returns 502 when backend is unreachable", async () => {
	vi.stubGlobal(
		"fetch",
		vi.fn().mockRejectedValue(new Error("connection refused")),
	);
	const result = await GET(new NextRequest(url), { params });
	expect(result.status).toBe(502);
	expect(await result.json()).toMatchObject({ code: "BACKEND_OFFLINE" });
});

it("preserves download bytes and filename", async () => {
	const bytes = Uint8Array.of(0, 255, 128, 10);
	const disposition = 'attachment; filename="report.bin"';
	vi.stubGlobal(
		"fetch",
		vi.fn().mockResolvedValue(
			new Response(bytes, {
				headers: {
					"content-type": "application/octet-stream",
					"content-disposition": disposition,
				},
			}),
		),
	);
	const result = await GET(new NextRequest(url), { params });
	expect(result.headers.get("content-disposition")).toBe(disposition);
	expect(new Uint8Array(await result.arrayBuffer())).toEqual(bytes);
});

it("forwards HEAD requests without a body and preserves safe headers", async () => {
	const upstream = new Response(null, {
		status: 200,
		headers: {
			"cache-control": "no-store",
			"x-request-id": "head-trace",
		},
	});
	const fetchMock = vi
		.fn<typeof globalThis.fetch>()
		.mockResolvedValue(upstream);
	vi.stubGlobal("fetch", fetchMock);
	const result = await HEAD(new NextRequest(url, { method: "HEAD" }), {
		params,
	});
	expect(fetchMock.mock.calls[0][1]?.method).toBe("HEAD");
	expect(fetchMock.mock.calls[0][1]?.body).toBeUndefined();
	expect(result.status).toBe(200);
	expect(result.headers.get("x-request-id")).toBe("head-trace");
	expect(result.headers.get("cache-control")).toBe("no-store");
});
