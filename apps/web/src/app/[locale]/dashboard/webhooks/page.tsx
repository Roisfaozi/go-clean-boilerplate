import { WebhooksProvider } from "./_components/webhooks-context";
import { WebhooksList } from "./_components/webhooks-list";
import { CreateWebhookDialog } from "./_components/create-webhook-dialog";
import { webhooksApi } from "~/lib/api/webhooks";
import { cookies } from "next/headers";

export default async function WebhooksPage() {
	const cookieStore = await cookies();
	const orgId = cookieStore.get("organization_id")?.value;

	let initialData = undefined;
	if (orgId) {
		try {
			initialData = await webhooksApi.list();
		} catch (error) {
			console.error("Failed to fetch initial webhooks on server", error);
		}
	}

	return (
		<WebhooksProvider initialData={initialData}>
			<div className="space-y-6">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">Webhooks</h2>
					<p className="text-muted-foreground">
						Receive HTTP callbacks for events in your organization.
					</p>
				</div>

				<div className="grid gap-6">
					<CreateWebhookDialog />
					<WebhooksList />
				</div>
			</div>
		</WebhooksProvider>
	);
}
