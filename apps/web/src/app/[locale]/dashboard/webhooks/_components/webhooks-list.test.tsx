import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";
import { WebhookLogs } from "./webhooks-list";

const contextMock = vi.hoisted(() => ({
	getWebhookLogs: vi.fn(),
}));

vi.mock("./webhooks-context", () => ({
	useWebhooks: () => ({
		getWebhookLogs: contextMock.getWebhookLogs,
	}),
}));

const log = {
	id: "log-1",
	webhook_id: "w1",
	event_type: "user.created",
	payload: "{}",
	response_status_code: 200,
	response_body: "ok",
	execution_time: 12,
	error_message: "",
	retry_count: 0,
	created_at: 1700000000000,
};

describe("WebhookLogs", () => {
	it("renders delivery logs on success", async () => {
		contextMock.getWebhookLogs.mockResolvedValue([log]);
		const user = userEvent.setup();
		render(<WebhookLogs webhookId="w1" />);

		await user.click(
			screen.getByRole("button", { name: /load delivery logs/i }),
		);

		expect(await screen.findByText("user.created")).toBeTruthy();
		expect(screen.getByText("200")).toBeTruthy();
		expect(screen.getByText("ok")).toBeTruthy();
	});

	it("shows the empty state when there are no deliveries", async () => {
		contextMock.getWebhookLogs.mockResolvedValue([]);
		const user = userEvent.setup();
		render(<WebhookLogs webhookId="w1" />);

		await user.click(
			screen.getByRole("button", { name: /load delivery logs/i }),
		);

		expect(await screen.findByText("No deliveries yet.")).toBeTruthy();
	});

	it("renders an error state with toast and retry on failure", async () => {
		contextMock.getWebhookLogs.mockRejectedValue(new Error("network"));
		const errorSpy = vi.spyOn(toast, "error");
		const user = userEvent.setup();
		render(<WebhookLogs webhookId="w1" />);

		await user.click(
			screen.getByRole("button", { name: /load delivery logs/i }),
		);

		expect(
			await screen.findByText("Failed to load delivery logs."),
		).toBeTruthy();
		expect(errorSpy).toHaveBeenCalledWith("Failed to load delivery logs");
		expect(
			screen.queryByRole("button", { name: /load delivery logs/i }),
		).toBeNull();

		contextMock.getWebhookLogs.mockResolvedValue([log]);
		await user.click(screen.getByRole("button", { name: /^retry$/i }));

		expect(await screen.findByText("user.created")).toBeTruthy();
		expect(screen.queryByText("Failed to load delivery logs.")).toBeNull();
	});
});
