import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StrictMode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { VerifyEmailStatus } from "./verify-email-status";

const verifyEmailMock = vi.hoisted(() => vi.fn());

vi.mock("~/lib/api/auth", async (importOriginal) => {
	const original = await importOriginal<typeof import("~/lib/api/auth")>();
	return {
		...original,
		authApi: {
			...original.authApi,
			verifyEmail: verifyEmailMock,
		},
	};
});

describe("VerifyEmailStatus", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows an invalid-link state without a token", () => {
		render(<VerifyEmailStatus />);

		expect(screen.getByText("Invalid verification link")).toBeTruthy();
		expect(verifyEmailMock).not.toHaveBeenCalled();
	});

	it("verifies the token and shows success", async () => {
		verifyEmailMock.mockResolvedValue({ data: { message: "verified" } });
		render(
			<StrictMode>
				<VerifyEmailStatus token="verification-token" />
			</StrictMode>,
		);

		expect(screen.getByText("Verifying your email address...")).toBeTruthy();
		expect(await screen.findByText("Email verified")).toBeTruthy();
		expect(verifyEmailMock).toHaveBeenCalledWith("verification-token");
		expect(verifyEmailMock).toHaveBeenCalledTimes(1);
	});

	it("shows failure and retries verification", async () => {
		verifyEmailMock.mockRejectedValueOnce(new Error("expired"));
		const user = userEvent.setup();
		render(<VerifyEmailStatus token="verification-token" />);

		expect(await screen.findByText("Verification failed")).toBeTruthy();
		verifyEmailMock.mockResolvedValueOnce({ data: { message: "verified" } });
		await user.click(screen.getByRole("button", { name: "Try Again" }));

		expect(await screen.findByText("Email verified")).toBeTruthy();
		expect(verifyEmailMock).toHaveBeenCalledTimes(2);
	});
});
