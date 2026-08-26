import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ResetPasswordForm } from "./reset-password-form";

const mocks = vi.hoisted(() => ({
	resetPassword: vi.fn(),
	push: vi.fn(),
}));

vi.mock("next/navigation", () => ({
	useRouter: () => ({ push: mocks.push }),
}));

vi.mock("~/lib/api/auth", async (importOriginal) => {
	const original = await importOriginal<typeof import("~/lib/api/auth")>();
	return {
		...original,
		authApi: {
			...original.authApi,
			resetPassword: mocks.resetPassword,
		},
	};
});

describe("ResetPasswordForm", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows an invalid-link state without a token", () => {
		render(<ResetPasswordForm />);

		expect(screen.getByText("Invalid reset link")).toBeTruthy();
		expect(screen.queryByLabelText("New Password")).toBeNull();
	});

	it("rejects mismatched passwords without calling the api", async () => {
		const user = userEvent.setup();
		render(<ResetPasswordForm token="reset-token" />);

		await user.type(screen.getByLabelText("New Password"), "password-one");
		await user.type(
			screen.getByLabelText("Confirm New Password"),
			"password-two",
		);
		await user.click(screen.getByRole("button", { name: "Update Password" }));

		expect(await screen.findByText("Passwords do not match.")).toBeTruthy();
		expect(mocks.resetPassword).not.toHaveBeenCalled();
	});

	it("submits the new password and redirects to login", async () => {
		mocks.resetPassword.mockResolvedValue({ data: { message: "updated" } });
		const user = userEvent.setup();
		render(<ResetPasswordForm token="reset-token" />);

		await user.type(screen.getByLabelText("New Password"), "new-password");
		await user.type(
			screen.getByLabelText("Confirm New Password"),
			"new-password",
		);
		await user.click(screen.getByRole("button", { name: "Update Password" }));

		expect(mocks.resetPassword).toHaveBeenCalledWith(
			"reset-token",
			"new-password",
		);
		expect(mocks.push).toHaveBeenCalledWith("/login");
	});
});
