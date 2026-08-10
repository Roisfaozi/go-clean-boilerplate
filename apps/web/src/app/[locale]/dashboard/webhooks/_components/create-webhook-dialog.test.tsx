import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { CreateWebhookDialog } from "./create-webhook-dialog";

const contextMock = vi.hoisted(() => ({
	createWebhook: vi.fn(),
}));

vi.mock("./webhooks-context", () => ({
	useWebhooks: () => ({
		createWebhook: contextMock.createWebhook,
	}),
}));

describe("CreateWebhookDialog", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	async function openDialog(user: ReturnType<typeof userEvent.setup>) {
		await user.click(screen.getByRole("button", { name: /create a webhook/i }));
		await screen.findByRole("dialog");
	}

	it("submits the form with parsed events", async () => {
		contextMock.createWebhook.mockResolvedValue({});
		const user = userEvent.setup();
		render(<CreateWebhookDialog />);
		await openDialog(user);

		await user.type(
			screen.getByPlaceholderText("Deployment notifier"),
			"Deploy notifier",
		);
		await user.type(
			screen.getByPlaceholderText("https://example.com/hooks/nexusos"),
			"https://example.com/hook",
		);
		await user.type(
			screen.getByPlaceholderText("At least 8 characters"),
			"supersecret",
		);
		await user.click(screen.getByRole("button", { name: "user.created" }));
		await user.click(screen.getByRole("button", { name: /create webhook$/i }));

		expect(contextMock.createWebhook).toHaveBeenCalledWith({
			name: "Deploy notifier",
			url: "https://example.com/hook",
			events: ["user.created"],
			secret: "supersecret",
		});
	});

	it("dedupes preset chips when clicked twice", async () => {
		const user = userEvent.setup();
		render(<CreateWebhookDialog />);
		await openDialog(user);

		const textarea = screen.getByPlaceholderText(
			"user.created, project.updated",
		);
		await user.click(screen.getByRole("button", { name: "user.created" }));
		await user.click(screen.getByRole("button", { name: "user.created" }));

		expect((textarea as HTMLTextAreaElement).value).toBe("user.created");
	});

	it("blocks submit and shows validation messages for an empty form", async () => {
		const user = userEvent.setup();
		render(<CreateWebhookDialog />);
		await openDialog(user);

		await user.click(screen.getByRole("button", { name: /create webhook$/i }));

		expect(
			await screen.findByText("Name must be at least 3 characters."),
		).toBeTruthy();
		expect(screen.getByText("Enter a valid URL.")).toBeTruthy();
		expect(screen.getByText("Enter at least one event type.")).toBeTruthy();
		expect(
			screen.getByText("Secret must be at least 8 characters."),
		).toBeTruthy();
		expect(contextMock.createWebhook).not.toHaveBeenCalled();
	});
});
