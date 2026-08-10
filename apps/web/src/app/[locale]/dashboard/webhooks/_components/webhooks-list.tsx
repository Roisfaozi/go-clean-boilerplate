"use client";

import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger,
	Badge,
	Button,
	Switch,
} from "@casbin/ui";
import { useState } from "react";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import type { Webhook, WebhookLog } from "~/lib/api/webhooks";
import { useWebhooks } from "./webhooks-context";

function formatTime(millis: number) {
	return new Date(millis).toLocaleString();
}

export function WebhookLogs({ webhookId }: { webhookId: string }) {
	const { getWebhookLogs } = useWebhooks();
	const [logs, setLogs] = useState<WebhookLog[] | null>(null);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState(false);

	async function loadLogs() {
		if (logs !== null) return;
		setLoading(true);
		setError(false);
		try {
			setLogs(await getWebhookLogs(webhookId));
		} catch (_error) {
			setError(true);
			toast.error("Failed to load delivery logs");
		} finally {
			setLoading(false);
		}
	}

	return (
		<div className="space-y-2">
			{loading && (
				<p className="text-muted-foreground text-sm">Loading logs...</p>
			)}
			{!loading && error && (
				<div className="flex items-center gap-2">
					<p className="text-destructive text-sm">
						Failed to load delivery logs.
					</p>
					<Button
						variant="outline"
						size="sm"
						onClick={() => {
							setError(false);
							loadLogs();
						}}
					>
						Retry
					</Button>
				</div>
			)}
			{!loading && !error && logs !== null && logs.length === 0 && (
				<p className="text-muted-foreground text-sm italic">
					No deliveries yet.
				</p>
			)}
			{!loading &&
				logs?.map((log) => (
					<div
						key={log.id}
						className="bg-muted/50 flex flex-col gap-1 rounded-md p-3 text-sm"
					>
						<div className="flex items-center gap-2">
							<Badge
								variant={
									log.response_status_code >= 200 &&
									log.response_status_code < 300
										? "default"
										: "destructive"
								}
								className="font-mono"
							>
								{log.response_status_code || "n/a"}
							</Badge>
							<span className="font-mono text-xs">{log.event_type}</span>
							<span className="text-muted-foreground ml-auto text-xs">
								{formatTime(log.created_at)}
							</span>
						</div>
						{log.error_message && (
							<p className="text-destructive text-xs">{log.error_message}</p>
						)}
						{log.response_body && (
							<pre className="bg-background max-h-32 overflow-auto rounded p-2 text-xs">
								{log.response_body}
							</pre>
						)}
					</div>
				))}
			{logs === null && !loading && !error && (
				<Button variant="outline" size="sm" onClick={loadLogs}>
					Load delivery logs
				</Button>
			)}
		</div>
	);
}

export function WebhooksList() {
	const { webhooks, isLoading, updateWebhook, deleteWebhook } = useWebhooks();

	if (isLoading && webhooks.length === 0) {
		return <div className="py-12 text-center">Loading webhooks...</div>;
	}

	if (webhooks.length === 0) {
		return (
			<div className="py-12 text-center">
				<p className="text-muted-foreground italic">
					No webhooks configured yet.
				</p>
			</div>
		);
	}

	return (
		<div className="bg-card rounded-md border">
			<Accordion multiple className="w-full">
				{webhooks.map((webhook) => (
					<AccordionItem key={webhook.id} value={webhook.id} className="px-6">
						<AccordionTrigger className="hover:no-underline">
							<div className="flex items-center gap-3">
								<span className="text-lg font-semibold">{webhook.name}</span>
								<Badge variant={webhook.is_active ? "default" : "secondary"}>
									{webhook.is_active ? "Active" : "Paused"}
								</Badge>
							</div>
						</AccordionTrigger>
						<AccordionContent>
							<div className="flex flex-col gap-3 pb-4">
								<div className="flex items-center gap-2">
									<code className="bg-muted rounded px-2 py-1 font-mono text-xs break-all">
										{webhook.url}
									</code>
								</div>
								<div className="flex flex-wrap gap-1.5">
									{webhook.events.map((event) => (
										<Badge
											key={event}
											variant="outline"
											className="font-mono text-[10px]"
										>
											{event}
										</Badge>
									))}
								</div>
								<div className="flex items-center gap-4 pt-2">
									<label className="flex cursor-pointer items-center gap-2 text-sm">
										<Switch
											checked={webhook.is_active}
											onCheckedChange={(checked) =>
												updateWebhook(webhook.id, { is_active: checked })
											}
										/>
										Active
									</label>
									<AlertDialog>
										<AlertDialogTrigger
											nativeButton={false}
											render={
												<Button
													variant="ghost"
													size="sm"
													className="text-destructive"
												>
													<Icon name="Trash" className="h-4 w-4" />
												</Button>
											}
										/>
										<AlertDialogContent>
											<AlertDialogHeader>
												<AlertDialogTitle>Delete webhook?</AlertDialogTitle>
												<AlertDialogDescription>
													This permanently deletes{" "}
													<strong>{webhook.name}</strong> and its delivery logs.
													This action cannot be undone.
												</AlertDialogDescription>
											</AlertDialogHeader>
											<AlertDialogFooter>
												<AlertDialogCancel>Cancel</AlertDialogCancel>
												<AlertDialogAction
													onClick={() => deleteWebhook(webhook.id)}
												>
													Delete
												</AlertDialogAction>
											</AlertDialogFooter>
										</AlertDialogContent>
									</AlertDialog>
								</div>
								<WebhookLogs webhookId={webhook.id} />
							</div>
						</AccordionContent>
					</AccordionItem>
				))}
			</Accordion>
		</div>
	);
}
