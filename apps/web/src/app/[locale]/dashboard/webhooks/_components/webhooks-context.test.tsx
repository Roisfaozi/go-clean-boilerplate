import { act, render } from "@testing-library/react";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";
import { useWebhooks } from "./webhooks-context";
import { WebhooksProvider } from "./webhooks-context";

const mocks = vi.hoisted(() => ({
	list: vi.fn(),
	create: vi.fn(),
	update: vi.fn(),
	delete: vi.fn(),
	logs: vi.fn(),
}));

const storeMock = vi.hoisted(() => ({
	currentOrganization: { id: "org-1" },
}));

vi.mock("~/lib/api/webhooks", () => ({
	webhooksApi: {
		list: mocks.list,
		create: mocks.create,
		update: mocks.update,
		delete: mocks.delete,
		logs: mocks.logs,
	},
}));

vi.mock("~/stores/use-organization-store", () => ({
	useOrganizationStore: () => ({
		currentOrganization: storeMock.currentOrganization,
	}),
}));

const createPayload = {
	name: "Deploy notifier",
	url: "https://example.com/hook",
	events: ["user.created"],
	secret: "supersecret",
};

describe("WebhooksProvider", () => {
	let ctx: ReturnType<typeof useWebhooks>;

	function renderProvider() {
		function Capture() {
			ctx = useWebhooks();
			return null;
		}
		return render(
			<WebhooksProvider>
				<Capture />
			</WebhooksProvider>,
		);
	}

	it("lists webhooks from the api", async () => {
		mocks.list.mockResolvedValue({ data: [] });
		renderProvider();
		await act(async () => {});
		expect(mocks.list).toHaveBeenCalled();
	});

	it("creates a webhook and revalidates the list", async () => {
		mocks.list.mockResolvedValue({ data: [] });
		mocks.create.mockResolvedValue({ data: { id: "w1" } });
		renderProvider();
		await act(async () => {});
		const successSpy = vi.spyOn(toast, "success");

		await act(async () => {
			await ctx.createWebhook(createPayload);
		});

		expect(mocks.create).toHaveBeenCalledWith(createPayload);
		expect(successSpy).toHaveBeenCalledWith("Webhook created successfully");
		expect(mocks.list.mock.calls.length).toBeGreaterThanOrEqual(2);
	});

	it("surfaces create errors through toast and rethrows", async () => {
		mocks.list.mockResolvedValue({ data: [] });
		mocks.create.mockRejectedValue(new Error("boom"));
		renderProvider();
		await act(async () => {});
		const errorSpy = vi.spyOn(toast, "error");

		await expect(
			act(async () => {
				await ctx.createWebhook(createPayload);
			}),
		).rejects.toThrow("boom");

		expect(errorSpy).toHaveBeenCalledWith("Failed to create webhook");
	});

	it("updates a webhook with the partial payload", async () => {
		mocks.list.mockResolvedValue({ data: [] });
		mocks.update.mockResolvedValue({ data: { id: "w1" } });
		renderProvider();
		await act(async () => {});

		await act(async () => {
			await ctx.updateWebhook("w1", { is_active: false });
		});

		expect(mocks.update).toHaveBeenCalledWith("w1", { is_active: false });
	});

	it("deletes a webhook", async () => {
		mocks.list.mockResolvedValue({ data: [] });
		mocks.delete.mockResolvedValue(undefined);
		renderProvider();
		await act(async () => {});

		await act(async () => {
			await ctx.deleteWebhook("w1");
		});

		expect(mocks.delete).toHaveBeenCalledWith("w1");
	});

	it("propagates log fetch errors to the caller", async () => {
		mocks.list.mockResolvedValue({ data: [] });
		mocks.logs.mockRejectedValue(new Error("network"));
		renderProvider();
		await act(async () => {});

		await expect(
			act(async () => {
				await ctx.getWebhookLogs("w1");
			}),
		).rejects.toThrow("network");
	});
});
