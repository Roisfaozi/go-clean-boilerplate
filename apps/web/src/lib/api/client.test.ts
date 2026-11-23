import { afterEach, expect, it, vi } from "vitest";
import { api } from "./client";

afterEach(() => {
	document.cookie = "csrf_token=; Max-Age=0; Path=/";
	vi.unstubAllGlobals();
});

it.each([
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"GET",
])("sets CSRF for %s according to method", async (method) => {
	document.cookie = "csrf_token=first; Path=/";
	const fetch = vi
		.fn<typeof globalThis.fetch>()
		.mockResolvedValue(Response.json({}));
	vi.stubGlobal("fetch", fetch);
	await api.request("/projects", { method });
	expect(new Headers(fetch.mock.calls[0][1]?.headers).get("X-CSRF-Token")).toBe(
		method === "GET" ? null : "first",
	);
});

it("refreshes on 401 and retries with the rotated token", async () => {
	document.cookie = "csrf_token=first; Path=/";
	const calls: unknown[] = [];
	vi.stubGlobal(
		"fetch",
		vi.fn<typeof fetch>(async (url, options) => {
			calls.push([
				url,
				options?.method,
				new Headers(options?.headers).get("X-CSRF-Token"),
			]);
			if (calls.length === 1) return new Response(null, { status: 401 });
			if (url === "/api/v1/auth/refresh")
				document.cookie = "csrf_token=second; Path=/";
			return Response.json({ ok: true });
		}),
	);
	expect(await api.put("/projects/p", {})).toEqual({ ok: true });
	expect(calls).toEqual([
		["/api/v1/projects/p", "PUT", "first"],
		["/api/v1/auth/refresh", "POST", "first"],
		["/api/v1/projects/p", "PUT", "second"],
	]);
});
