import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ForgotPasswordForm } from "./forgot-password-form";

const forgotPasswordMock = vi.hoisted(() => vi.fn());

vi.mock("~/lib/api/auth", async (importOriginal) => {
	const original = await importOriginal<typeof import("~/lib/api/auth")>();
	return {
		...original,
		authApi: {
			...original.authApi,
			forgotPassword: forgotPasswordMock,
		},
	};
});

describe("ForgotPasswordForm", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("rejects an empty email without calling the api", async () => {
		const user = userEvent.setup();
		render(<ForgotPasswordForm />);

		await user.click(screen.getByRole("button", { name: "Send Reset Link" }));

		expect(
			await screen.findByText("Please enter a valid email address."),
		).toBeTruthy();
		expect(forgotPasswordMock).not.toHaveBeenCalled();
	});

	it("submits the email and shows the enumeration-safe success state", async () => {
		forgotPasswordMock.mockResolvedValue({ data: { message: "sent" } });
		const user = userEvent.setup();
		render(<ForgotPasswordForm />);

		await user.type(screen.getByLabelText("Email"), "user@example.com");
		await user.click(screen.getByRole("button", { name: "Send Reset Link" }));

		expect(forgotPasswordMock).toHaveBeenCalledWith({
			email: "user@example.com",
		});
		expect(await screen.findByText("Check your inbox")).toBeTruthy();
		expect(screen.getByText(/If the email is registered/)).toBeTruthy();
	});
});
