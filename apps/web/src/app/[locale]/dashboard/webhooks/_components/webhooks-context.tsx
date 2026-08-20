"use client";

import { createContext, useContext, useCallback, type ReactNode } from "react";
import { type Webhook, type WebhookLog, webhooksApi } from "~/lib/api/webhooks";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";
import useSWR from "swr";

interface WebhooksContextType {
	webhooks: Webhook[];
	isLoading: boolean;
	fetchWebhooks: () => Promise<void>;
	createWebhook: (data: {
		name: string;
		url: string;
		events: string[];
		secret: string;
	}) => Promise<void>;
	updateWebhook: (
		id: string,
		data: {
			name?: string;
			url?: string;
			events?: string[];
			secret?: string;
			is_active?: boolean;
		},
	) => Promise<void>;
	deleteWebhook: (id: string) => Promise<void>;
	getWebhookLogs: (id: string) => Promise<WebhookLog[]>;
}

const WebhooksContext = createContext<WebhooksContextType | undefined>(
	undefined,
);

export function WebhooksProvider({
	children,
	initialData,
}: {
	children: ReactNode;
	initialData?: Webhook[];
}) {
	const { currentOrganization } = useOrganizationStore();

	const {
		data: webhooks = [],
		isLoading,
		mutate,
	} = useSWR(
		currentOrganization ? ["/api/v1/webhooks", currentOrganization.id] : null,
		() => webhooksApi.list(),
		{ fallbackData: initialData, keepPreviousData: true },
	);

	const fetchWebhooks = useCallback(async () => {
		await mutate();
	}, [mutate]);

	const createWebhook = useCallback(
		async (data: {
			name: string;
			url: string;
			events: string[];
			secret: string;
		}) => {
			if (!currentOrganization) return;
			try {
				await webhooksApi.create(data);
				toast.success("Webhook created successfully");
				await mutate();
			} catch (error) {
				toast.error("Failed to create webhook");
				throw error;
			}
		},
		[currentOrganization, mutate],
	);

	const updateWebhook = useCallback(
		async (
			id: string,
			data: {
				name?: string;
				url?: string;
				events?: string[];
				secret?: string;
				is_active?: boolean;
			},
		) => {
			if (!currentOrganization) return;
			try {
				await webhooksApi.update(id, data);
				toast.success("Webhook updated successfully");
				await mutate();
			} catch (_error) {
				toast.error("Failed to update webhook");
			}
		},
		[currentOrganization, mutate],
	);

	const deleteWebhook = useCallback(
		async (id: string) => {
			if (!currentOrganization) return;
			try {
				await webhooksApi.delete(id);
				toast.success("Webhook deleted");
				await mutate();
			} catch (_error) {
				toast.error("Failed to delete webhook");
			}
		},
		[currentOrganization, mutate],
	);

	const getWebhookLogs = useCallback(async (id: string) => {
		return webhooksApi.logs(id);
	}, []);

	return (
		<WebhooksContext.Provider
			value={{
				webhooks,
				isLoading,
				fetchWebhooks,
				createWebhook,
				updateWebhook,
				deleteWebhook,
				getWebhookLogs,
			}}
		>
			{children}
		</WebhooksContext.Provider>
	);
}

export function useWebhooks() {
	const context = useContext(WebhooksContext);
	if (context === undefined) {
		throw new Error("useWebhooks must be used within a WebhooksProvider");
	}
	return context;
}
