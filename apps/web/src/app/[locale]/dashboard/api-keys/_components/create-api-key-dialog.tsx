"use client";

import {
	Badge,
	Button,
	Card,
	Checkbox,
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
} from "@casbin/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";
import { Icon } from "~/components/shared/icon";
import type { ApiKeyCreated } from "~/lib/api/api-keys";
import { useApiKeys } from "./api-keys-context";

// Scopes the backend actually enforces for API keys (tenantAuthorized and
// authorized route groups in internal/router/router.go). Scopes for
// user/permission/access-right/endpoint/audit are not offered here because
// those routes are JWT-only or require admin:manage instead.
const API_KEY_SCOPES: { resource: string; actions: string[] }[] = [
	{ resource: "org", actions: ["view", "manage"] },
	{ resource: "project", actions: ["view", "manage"] },
	{ resource: "role", actions: ["view", "manage"] },
	{ resource: "member", actions: ["manage"] },
	{ resource: "presence", actions: ["view"] },
	{ resource: "webhook", actions: ["manage"] },
	{ resource: "admin", actions: ["manage"] },
];

const apiKeySchema = z.object({
	name: z.string().min(3, { message: "Name must be at least 3 characters." }),
	scopes: z.array(z.string()),
	expires_at: z.string().optional(),
});

type ApiKeyFormValues = z.infer<typeof apiKeySchema>;

export function CreateApiKeyDialog() {
	const [isOpen, setIsOpen] = useState(false);
	const [createdKey, setCreatedKey] = useState<ApiKeyCreated | null>(null);
	const [copied, setCopied] = useState(false);
	const { createApiKey } = useApiKeys();

	const form = useForm<ApiKeyFormValues>({
		resolver: zodResolver(apiKeySchema),
		defaultValues: {
			name: "",
			scopes: [],
			expires_at: undefined,
		},
	});

	const selectedScopes = form.watch("scopes");

	function toggleScope(scope: string, checked: boolean) {
		const next = checked
			? [...selectedScopes, scope]
			: selectedScopes.filter((s) => s !== scope);
		form.setValue("scopes", next, { shouldDirty: true });
	}

	async function onSubmit(values: ApiKeyFormValues) {
		try {
			const created = await createApiKey({
				name: values.name,
				scopes: values.scopes,
				expires_at: values.expires_at
					? new Date(values.expires_at).toISOString()
					: null,
			});
			setCreatedKey(created);
			form.reset();
		} catch (_error) {
			// Toast already shown by context
		}
	}

	function close() {
		setIsOpen(false);
		setCreatedKey(null);
		setCopied(false);
	}

	function copyKey() {
		if (!createdKey) return;
		navigator.clipboard.writeText(createdKey.api_key).then(() => {
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		});
	}

	return (
		<Dialog
			open={isOpen}
			onOpenChange={(open) => (open ? setIsOpen(true) : close())}
		>
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
						<p className="text-xl font-semibold">Create an API key</p>
						<p className="text-muted-foreground text-sm">
							Generate a new key for programmatic access
						</p>
					</Card>
				}
			/>
			<DialogContent className="sm:max-w-[520px]">
				{createdKey ? (
					<>
						<DialogHeader>
							<DialogTitle>API key created</DialogTitle>
							<DialogDescription>
								Copy the key now. For security reasons, it cannot be shown
								again.
							</DialogDescription>
						</DialogHeader>
						<div className="space-y-4 py-4">
							<div className="flex items-center gap-2">
								<code className="bg-muted flex-1 rounded-md px-3 py-2 text-sm break-all">
									{createdKey.api_key}
								</code>
								<Button
									type="button"
									variant="outline"
									size="icon"
									onClick={copyKey}
								>
									{copied ? (
										<Icon name="Check" className="h-4 w-4" />
									) : (
										<Icon name="Copy" className="h-4 w-4" />
									)}
								</Button>
							</div>
							<p className="text-muted-foreground text-xs">
								Scopes: {createdKey.scopes.join(", ") || "none"}
							</p>
						</div>
						<DialogFooter>
							<Button onClick={close} className="w-full">
								Done
							</Button>
						</DialogFooter>
					</>
				) : (
					<>
						<DialogHeader>
							<DialogTitle>Create API key</DialogTitle>
							<DialogDescription>
								Choose a name and the scopes this key is allowed to use.
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
											<FormLabel>Key Name</FormLabel>
											<FormControl>
												<Input placeholder="CI/CD token" {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<FormField
									control={form.control}
									name="expires_at"
									render={({ field }) => (
										<FormItem>
											<FormLabel>Expires At (optional)</FormLabel>
											<FormControl>
												<Input type="datetime-local" {...field} />
											</FormControl>
											<FormMessage />
										</FormItem>
									)}
								/>
								<div>
									<FormLabel>Scopes</FormLabel>
									<div className="bg-muted/50 mt-2 max-h-52 space-y-3 overflow-y-auto rounded-md p-3">
										{API_KEY_SCOPES.map((group) => (
											<div key={group.resource}>
												<p className="text-muted-foreground mb-1.5 text-xs font-medium uppercase tracking-wide">
													{group.resource}
												</p>
												<div className="flex flex-wrap gap-x-4 gap-y-1.5">
													{group.actions.map((action) => {
														const scope = `${group.resource}:${action}`;
														return (
															<label
																key={scope}
																className="flex items-center gap-1.5 text-sm"
															>
																<Checkbox
																	checked={selectedScopes.includes(scope)}
																	onCheckedChange={(checked) =>
																		toggleScope(scope, checked === true)
																	}
																/>
																<span className="font-mono text-xs">
																	{scope}
																</span>
															</label>
														);
													})}
												</div>
											</div>
										))}
									</div>
									{selectedScopes.length === 0 && (
										<p className="text-muted-foreground mt-1.5 text-xs">
											No scopes selected. The key will not be usable for
											protected endpoints.
										</p>
									)}
									<p className="text-muted-foreground mt-1.5 text-xs">
										admin:manage unlocks admin endpoints (permissions, audit,
										access rights, users, organizations). Member and presence
										endpoints also require org:view.
									</p>
								</div>
								<DialogFooter className="pt-4">
									<Button type="submit" className="w-full">
										<Icon name="Plus" className="mr-2 h-4 w-4" />
										Create API key
									</Button>
								</DialogFooter>
							</form>
						</Form>
					</>
				)}
			</DialogContent>
		</Dialog>
	);
}
