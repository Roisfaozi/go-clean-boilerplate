import { act, render } from "@testing-library/react";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";
import type { ApiKeyCreated } from "~/lib/api/api-keys";
import { ApiKeysProvider } from "./api-keys-context";
import { useApiKeys } from "./api-keys-context";

const mocks = vi.hoisted(() => ({
	list: vi.fn(),
	create: vi.fn(),
	revoke: vi.fn(),
}));

const storeMock = vi.hoisted<{ currentOrganization: { id: string } | null }>(
	() => ({
		currentOrganization: { id: "org-1" },
	}),
);

vi.mock("~/lib/api/api-keys", () => ({
	apiKeysApi: {
		list: mocks.list,
		create: mocks.create,
		revoke: mocks.revoke,
	},
}));

vi.mock("~/stores/use-organization-store", () => ({
	useOrganizationStore: () => ({
		currentOrganization: storeMock.currentOrganization,
	}),
}));

describe("ApiKeysProvider", () => {
	let ctx: ReturnType<typeof useApiKeys>;

	function renderProvider() {
		function Capture() {
			ctx = useApiKeys();
			return null;
		}
		return render(
			<ApiKeysProvider>
				<Capture />
			</ApiKeysProvider>,
		);
	}

	it("creates an api key and returns the created key with raw value", async () => {
		mocks.list.mockResolvedValue({ data: [] });
		const created = {
			id: "k1",
			name: "CI/CD token",
			organization_id: "org-1",
			user_id: "u1",
			scopes: ["project:view"],
			expires_at: null,
			last_used_at: null,
			is_active: true,
			created_at: "2026-01-01T00:00:00Z",
			api_key: "sk-live-123",
		};
		mocks.create.mockResolvedValue(created);
		renderProvider();
		await act(async () => {});
		const successSpy = vi.spyOn(toast, "success");

		let result: ApiKeyCreated | undefined;
		await act(async () => {
			result = await ctx.createApiKey({
				name: "CI/CD token",
				scopes: ["project:view"],
				expires_at: null,
			});
		});

		expect(mocks.create).toHaveBeenCalledWith({
			name: "CI/CD token",
			scopes: ["project:view"],
			expires_at: null,
		});
		expect(result).toEqual(created);
		expect(successSpy).toHaveBeenCalledWith("API key created successfully");
	});

	it("throws when no organization is selected", async () => {
		storeMock.currentOrganization = null;
		mocks.list.mockResolvedValue({ data: [] });
		renderProvider();
		await act(async () => {});

		await expect(
			act(async () => {
				await ctx.createApiKey({ name: "x", scopes: [] });
			}),
		).rejects.toThrow("Organization context required");
	});

	it("revokes an api key", async () => {
		storeMock.currentOrganization = { id: "org-1" };
		mocks.list.mockResolvedValue({ data: [] });
		mocks.revoke.mockResolvedValue(undefined);
		renderProvider();
		await act(async () => {});

		await act(async () => {
			await ctx.revokeApiKey("k1");
		});

		expect(mocks.revoke).toHaveBeenCalledWith("k1");
	});

	it("toasts on revoke failure without rethrowing", async () => {
		storeMock.currentOrganization = { id: "org-1" };
		mocks.list.mockResolvedValue({ data: [] });
		mocks.revoke.mockRejectedValue(new Error("boom"));
		renderProvider();
		await act(async () => {});
		const errorSpy = vi.spyOn(toast, "error");

		await act(async () => {
			await ctx.revokeApiKey("k1");
		});

		expect(errorSpy).toHaveBeenCalledWith("Failed to revoke API key");
	});
});
