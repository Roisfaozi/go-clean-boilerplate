"use client";

import {
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
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@casbin/ui";
import { Icon } from "~/components/shared/icon";
import { useApiKeys } from "./api-keys-context";

function formatDate(value: string | null) {
	if (!value) return "Never";
	return new Date(value).toLocaleString("en-US", { timeZone: "Asia/Jakarta" });
}

export function ApiKeysList() {
	const { apiKeys, isLoading, revokeApiKey } = useApiKeys();

	if (isLoading && apiKeys.length === 0) {
		return <div className="py-12 text-center">Loading API keys...</div>;
	}

	if (apiKeys.length === 0) {
		return (
			<div className="py-12 text-center">
				<p className="text-muted-foreground italic">
					No API keys yet. Create one to get started.
				</p>
			</div>
		);
	}

	return (
		<div className="bg-card rounded-md border">
			<Table>
				<TableHeader>
					<TableRow>
						<TableHead>Name</TableHead>
						<TableHead>Scopes</TableHead>
						<TableHead>Expires At</TableHead>
						<TableHead>Last Used</TableHead>
						<TableHead>Status</TableHead>
						<TableHead className="text-right">Actions</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{apiKeys.map((key) => (
						<TableRow key={key.id}>
							<TableCell className="font-medium">{key.name}</TableCell>
							<TableCell>
								<div className="flex max-w-[280px] flex-wrap gap-1">
									{key.scopes.length === 0 && (
										<span className="text-muted-foreground text-xs italic">
											no scopes
										</span>
									)}
									{key.scopes.slice(0, 4).map((scope) => (
										<Badge
											key={scope}
											variant="outline"
											className="font-mono text-[10px]"
										>
											{scope}
										</Badge>
									))}
									{key.scopes.length > 4 && (
										<Badge variant="outline" className="text-[10px]">
											+{key.scopes.length - 4}
										</Badge>
									)}
								</div>
							</TableCell>
							<TableCell>{formatDate(key.expires_at)}</TableCell>
							<TableCell>{formatDate(key.last_used_at)}</TableCell>
							<TableCell>
								<Badge variant={key.is_active ? "default" : "destructive"}>
									{key.is_active ? "Active" : "Revoked"}
								</Badge>
							</TableCell>
							<TableCell className="text-right">
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
											<AlertDialogTitle>Revoke API key?</AlertDialogTitle>
											<AlertDialogDescription>
												This permanently revokes <strong>{key.name}</strong>.
												This action cannot be undone.
											</AlertDialogDescription>
										</AlertDialogHeader>
										<AlertDialogFooter>
											<AlertDialogCancel>Cancel</AlertDialogCancel>
											<AlertDialogAction onClick={() => revokeApiKey(key.id)}>
												Revoke
											</AlertDialogAction>
										</AlertDialogFooter>
									</AlertDialogContent>
								</AlertDialog>
							</TableCell>
						</TableRow>
					))}
				</TableBody>
			</Table>
		</div>
	);
}
