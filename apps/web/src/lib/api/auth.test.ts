import { beforeEach, describe, expect, it, vi } from "vitest";
import { authApi } from "./auth";

const postMock = vi.hoisted(() => vi.fn());

vi.mock("./client", () => ({
	api: {
		post: postMock,
		silentGet: vi.fn(),
	},
}));

describe("authApi recovery methods", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("maps forgot password to the backend contract", async () => {
		postMock.mockResolvedValue({ data: { message: "sent" } });

		await authApi.forgotPassword({ email: "user@example.com" });

		expect(postMock).toHaveBeenCalledWith("/auth/forgot-password", {
			email: "user@example.com",
		});
	});

	it("maps reset password to token and new_password", async () => {
		postMock.mockResolvedValue({ data: { message: "updated" } });

		await authApi.resetPassword("reset-token", "new-password");

		expect(postMock).toHaveBeenCalledWith("/auth/reset-password", {
			token: "reset-token",
			new_password: "new-password",
		});
	});

	it("maps verify email to the backend contract", async () => {
		postMock.mockResolvedValue({ data: { message: "verified" } });

		await authApi.verifyEmail("verification-token");

		expect(postMock).toHaveBeenCalledWith("/auth/verify-email", {
			token: "verification-token",
		});
	});
});
