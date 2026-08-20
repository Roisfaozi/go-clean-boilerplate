import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { CreateApiKeyDialog } from "./create-api-key-dialog";

const contextMock = vi.hoisted(() => ({
	createApiKey: vi.fn(),
}));

vi.mock("./api-keys-context", () => ({
	useApiKeys: () => ({
		createApiKey: contextMock.createApiKey,
	}),
}));

describe("CreateApiKeyDialog", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	async function openDialog(user: ReturnType<typeof userEvent.setup>) {
		await user.click(
			screen.getByRole("button", { name: /create an api key/i }),
		);
		await screen.findByRole("dialog");
	}

	it("submits the form with selected scopes", async () => {
		contextMock.createApiKey.mockResolvedValue({
			id: "k1",
			name: "CI/CD token",
			organization_id: "org-1",
			user_id: "u1",
			scopes: ["project:manage"],
			expires_at: null,
			last_used_at: null,
			is_active: true,
			created_at: "2026-01-01T00:00:00Z",
			api_key: "sk-live-123",
		});
		const user = userEvent.setup();
		render(<CreateApiKeyDialog />);
		await openDialog(user);

		await user.type(screen.getByPlaceholderText("CI/CD token"), "CI/CD token");
		await user.click(screen.getByRole("checkbox", { name: "project:manage" }));
		await user.click(screen.getByRole("button", { name: /create api key$/i }));

		expect(contextMock.createApiKey).toHaveBeenCalledWith({
			name: "CI/CD token",
			scopes: ["project:manage"],
			expires_at: null,
		});
	});

	it("shows the raw key only once after creation and closes on Done", async () => {
		contextMock.createApiKey.mockResolvedValue({
			id: "k1",
			name: "CI/CD token",
			organization_id: "org-1",
			user_id: "u1",
			scopes: [],
			expires_at: null,
			last_used_at: null,
			is_active: true,
			created_at: "2026-01-01T00:00:00Z",
			api_key: "sk-live-123",
		});
		const user = userEvent.setup();
		render(<CreateApiKeyDialog />);
		await openDialog(user);

		await user.type(screen.getByPlaceholderText("CI/CD token"), "CI/CD token");
		await user.click(screen.getByRole("button", { name: /create api key$/i }));

		expect(await screen.findByText("sk-live-123")).toBeTruthy();
		expect(screen.getAllByText("sk-live-123")).toHaveLength(1);

		await user.click(screen.getByRole("button", { name: /^done$/i }));
		expect(screen.queryByRole("dialog")).toBeNull();
	});

	it("does not submit without a name", async () => {
		const user = userEvent.setup();
		render(<CreateApiKeyDialog />);
		await openDialog(user);

		await user.click(screen.getByRole("button", { name: /create api key$/i }));

		expect(
			await screen.findByText("Name must be at least 3 characters."),
		).toBeTruthy();
		expect(contextMock.createApiKey).not.toHaveBeenCalled();
	});
});
