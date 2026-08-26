"use client";

import { createContext, useContext, useCallback, type ReactNode } from "react";
import {
	type ApiKey,
	type ApiKeyCreated,
	apiKeysApi,
} from "~/lib/api/api-keys";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";
import useSWR from "swr";

interface ApiKeysContextType {
	apiKeys: ApiKey[];
	isLoading: boolean;
	fetchApiKeys: () => Promise<void>;
	createApiKey: (data: {
		name: string;
		scopes: string[];
		expires_at?: string | null;
	}) => Promise<ApiKeyCreated>;
	revokeApiKey: (id: string) => Promise<void>;
}

const ApiKeysContext = createContext<ApiKeysContextType | undefined>(undefined);

export function ApiKeysProvider({
	children,
	initialData,
}: {
	children: ReactNode;
	initialData?: ApiKey[];
}) {
	const { currentOrganization } = useOrganizationStore();

	const {
		data: apiKeys = [],
		isLoading,
		mutate,
	} = useSWR(
		currentOrganization ? ["/api/v1/api-keys", currentOrganization.id] : null,
		() => apiKeysApi.list(),
		{ fallbackData: initialData, keepPreviousData: true },
	);

	const fetchApiKeys = useCallback(async () => {
		await mutate();
	}, [mutate]);

	const createApiKey = useCallback(
		async (data: {
			name: string;
			scopes: string[];
			expires_at?: string | null;
		}) => {
			if (!currentOrganization) {
				throw new Error("Organization context required");
			}
			try {
				const created = await apiKeysApi.create(data);
				toast.success("API key created successfully");
				await mutate();
				return created;
			} catch (error) {
				toast.error("Failed to create API key");
				throw error;
			}
		},
		[currentOrganization, mutate],
	);

	const revokeApiKey = useCallback(
		async (id: string) => {
			if (!currentOrganization) return;
			try {
				await apiKeysApi.revoke(id);
				toast.success("API key revoked");
				await mutate();
			} catch (_error) {
				toast.error("Failed to revoke API key");
			}
		},
		[currentOrganization, mutate],
	);

	return (
		<ApiKeysContext.Provider
			value={{
				apiKeys,
				isLoading,
				fetchApiKeys,
				createApiKey,
				revokeApiKey,
			}}
		>
			{children}
		</ApiKeysContext.Provider>
	);
}

export function useApiKeys() {
	const context = useContext(ApiKeysContext);
	if (context === undefined) {
		throw new Error("useApiKeys must be used within an ApiKeysProvider");
	}
	return context;
}
