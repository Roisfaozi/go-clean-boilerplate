// @vitest-environment node
import { afterEach, expect, it, vi } from "vitest";
import { api } from "./client";

vi.mock("next/headers", () => ({
	cookies: async () => {
		const entries = [
			{ name: "access_token", value: "session" },
			{ name: "csrf_token", value: "token" },
			{ name: "organization_id", value: "org" },
		];
		return {
			getAll: () => entries,
			get: (name: string) => entries.find((cookie) => cookie.name === name),
		};
	},
}));

afterEach(() => vi.unstubAllGlobals());

it("pairs the SSR mutation header with the request's cookies and tenant", async () => {
	const fetch = vi
		.fn<typeof globalThis.fetch>()
		.mockResolvedValue(Response.json({}));
	vi.stubGlobal("fetch", fetch);
	await api.post("/projects", { name: "Test", domain: "example.test" });
	const headers = new Headers(fetch.mock.calls[0][1]?.headers);
	expect(headers.get("cookie")).toContain("csrf_token=token");
	expect(headers.get("x-csrf-token")).toBe("token");
	expect(headers.get("x-organization-id")).toBe("org");
});
