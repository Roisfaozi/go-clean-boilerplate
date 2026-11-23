// @vitest-environment node
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const cookieStoreMock = {
	set: vi.fn(),
	delete: vi.fn(),
};

vi.mock("next/headers", () => ({
	cookies: vi.fn(async () => cookieStoreMock),
}));

vi.mock("next/navigation", () => ({
	redirect: vi.fn(),
}));

import { loginAction } from "./auth";

describe("loginAction", () => {
	beforeEach(() => {
		cookieStoreMock.set.mockReset();
		cookieStoreMock.delete.mockReset();
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it("extracts and stores csrf_token, access_token, and refresh_token cookies", async () => {
		const upstreamHeaders = new Headers();
		upstreamHeaders.append(
			"set-cookie",
			"csrf_token=csrf-123; Path=/; SameSite=Lax",
		);
		upstreamHeaders.append(
			"set-cookie",
			"refresh_token=ref-456; HttpOnly; Path=/; SameSite=Lax",
		);

		vi.stubGlobal(
			"fetch",
			vi.fn().mockResolvedValue(
				new Response(
					JSON.stringify({
						data: {
							access_token: "acc-789",
							user: { id: "u1", username: "admin" },
						},
					}),
					{
						status: 200,
						headers: upstreamHeaders,
					},
				),
			),
		);

		const result = await loginAction({
			username: "admin",
			password: "Password0!",
		});

		expect(result?.data).toEqual({
			success: true,
			user: { id: "u1", username: "admin" },
		});

		expect(cookieStoreMock.set).toHaveBeenCalledTimes(3);
		expect(cookieStoreMock.set).toHaveBeenCalledWith(
			"csrf_token",
			"csrf-123",
			expect.objectContaining({ path: "/", sameSite: "lax" }),
		);
		const csrfCallOptions = cookieStoreMock.set.mock.calls.find(
			(call) => call[0] === "csrf_token",
		)?.[2];
		expect(csrfCallOptions?.httpOnly).not.toBe(true);

		expect(cookieStoreMock.set).toHaveBeenCalledWith(
			"access_token",
			"acc-789",
			expect.objectContaining({ httpOnly: true, path: "/", sameSite: "lax" }),
		);
		expect(cookieStoreMock.set).toHaveBeenCalledWith(
			"refresh_token",
			"ref-456",
			expect.objectContaining({ httpOnly: true, path: "/", sameSite: "lax" }),
		);
	});

	it("returns failure and does not store cookies when session cookies are missing", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn().mockResolvedValue(
				new Response(
					JSON.stringify({
						data: {
							access_token: "acc-789",
							user: { id: "u1", username: "admin" },
						},
					}),
					{ status: 200 },
				),
			),
		);

		const result = await loginAction({
			username: "admin",
			password: "Password0!",
		});

		expect(result?.data).toEqual({
			success: false,
			message: "Missing session cookies",
		});
		expect(cookieStoreMock.set).not.toHaveBeenCalled();
	});
});
