import { ApiKeysProvider } from "./_components/api-keys-context";
import { ApiKeysList } from "./_components/api-keys-list";
import { CreateApiKeyDialog } from "./_components/create-api-key-dialog";
import { apiKeysApi } from "~/lib/api/api-keys";
import { cookies } from "next/headers";

export default async function ApiKeysPage() {
	const cookieStore = await cookies();
	const orgId = cookieStore.get("organization_id")?.value;

	let initialData = undefined;
	if (orgId) {
		try {
			initialData = await apiKeysApi.list();
		} catch (error) {
			console.error("Failed to fetch initial API keys on server", error);
		}
	}

	return (
		<ApiKeysProvider initialData={initialData}>
			<div className="space-y-6">
				<div>
					<h2 className="text-2xl font-bold tracking-tight">API Keys</h2>
					<p className="text-muted-foreground">
						Manage keys for programmatic access to your organization.
					</p>
				</div>

				<div className="grid gap-6">
					<CreateApiKeyDialog />
					<ApiKeysList />
				</div>
			</div>
		</ApiKeysProvider>
	);
}
