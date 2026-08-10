"use client";

import {
	Badge,
	Button,
	Card,
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
	Input,
	Textarea,
} from "@casbin/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";
import { Icon } from "~/components/shared/icon";
import { useWebhooks } from "./webhooks-context";

const EVENT_PRESETS = [
	"user.created",
	"user.updated",
	"user.deleted",
	"project.created",
	"project.updated",
	"project.deleted",
	"organization.created",
	"organization.updated",
	"member.invited",
	"member.removed",
	"role.created",
	"role.deleted",
];

const webhookSchema = z.object({
	name: z.string().min(3, { message: "Name must be at least 3 characters." }),
	url: z.string().url({ message: "Enter a valid URL." }),
	events: z.string().min(1, { message: "Enter at least one event type." }),
	secret: z
		.string()
		.min(8, { message: "Secret must be at least 8 characters." }),
});

type WebhookFormValues = z.infer<typeof webhookSchema>;

function parseEvents(value: string): string[] {
	return value
		.split(",")
		.map((e) => e.trim())
		.filter(Boolean);
}

export function CreateWebhookDialog() {
	const [isOpen, setIsOpen] = useState(false);
	const { createWebhook } = useWebhooks();

	const form = useForm<WebhookFormValues>({
		resolver: zodResolver(webhookSchema),
		defaultValues: {
			name: "",
			url: "",
			events: "",
			secret: "",
		},
	});

	async function onSubmit(values: WebhookFormValues) {
		try {
			await createWebhook({
				name: values.name,
				url: values.url,
				events: parseEvents(values.events),
				secret: values.secret,
			});
			form.reset();
			setIsOpen(false);
		} catch (_error) {
			// Toast already shown by context
		}
	}

	function addPreset(event: string) {
		const current = parseEvents(form.getValues("events"));
		if (!current.includes(event)) {
			form.setValue("events", [...current, event].join(", "), {
				shouldDirty: true,
			});
		}
	}

	return (
		<Dialog open={isOpen} onOpenChange={setIsOpen}>
			<DialogTrigger
				nativeButton={false}
				render={
					<Card
						role="button"
						className="hover:bg-accent flex flex-col items-center justify-center gap-y-2.5 p-8 text-center transition-colors"
					>
						<div className="bg-primary/10 text-primary rounded-full p-3">
							<Icon name="Plus" className="h-8 w-8" />
						</div>
						<p className="text-xl font-semibold">Create a webhook</p>
						<p className="text-muted-foreground text-sm">
							Notify an external URL on events
						</p>
					</Card>
				}
			/>
			<DialogContent className="sm:max-w-[520px]">
				<DialogHeader>
					<DialogTitle>Create Webhook</DialogTitle>
					<DialogDescription>
						Deliver HTTP callbacks to your endpoint when events occur in this
						organization.
					</DialogDescription>
				</DialogHeader>
				<Form {...form}>
					<form
						onSubmit={form.handleSubmit(onSubmit)}
						className="space-y-4 py-4"
					>
						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Name</FormLabel>
									<FormControl>
										<Input placeholder="Deployment notifier" {...field} />
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>
						<FormField
							control={form.control}
							name="url"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Endpoint URL</FormLabel>
									<FormControl>
										<Input
											placeholder="https://example.com/hooks/nexusos"
											type="url"
											{...field}
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>
						<FormField
							control={form.control}
							name="secret"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Secret</FormLabel>
									<FormControl>
										<Input
											placeholder="At least 8 characters"
											type="password"
											{...field}
										/>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>
						<FormField
							control={form.control}
							name="events"
							render={({ field }) => (
								<FormItem>
									<FormLabel>Events</FormLabel>
									<FormControl>
										<Textarea
											placeholder="user.created, project.updated"
											{...field}
										/>
									</FormControl>
									<div className="mt-2 flex flex-wrap gap-1.5">
										{EVENT_PRESETS.map((event) => (
											<button
												key={event}
												type="button"
												onClick={() => addPreset(event)}
												className="hover:bg-primary/10 focus-visible:ring-ring rounded-full border px-2 py-0.5 text-xs transition-colors focus-visible:ring-2"
											>
												{event}
											</button>
										))}
									</div>
									<FormMessage />
								</FormItem>
							)}
						/>
						<DialogFooter className="pt-4">
							<Button type="submit" className="w-full">
								<Icon name="Plus" className="mr-2 h-4 w-4" />
								Create Webhook
							</Button>
						</DialogFooter>
					</form>
				</Form>
			</DialogContent>
		</Dialog>
	);
}
