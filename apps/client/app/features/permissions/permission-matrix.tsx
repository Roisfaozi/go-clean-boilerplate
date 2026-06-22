import { useState, useEffect } from "react";
import { cn } from "@casbin/ui";
import { Checkbox } from "@casbin/ui";
import { NexusCard } from "@casbin/ui";
import { Badge } from "@casbin/ui";
import { useToast } from "@casbin/ui";

interface PermissionMatrixProps {
	roles: string[];
	resources: string[];
	actions: string[];
	initialPermissions?: Record<string, Record<string, string[]>>;
	onPermissionChange?: (
		role: string,
		resource: string,
		action: string,
		granted: boolean,
	) => void;
}

export function PermissionMatrix({
	roles,
	resources,
	actions,
	initialPermissions = {},
	onPermissionChange,
}: PermissionMatrixProps) {
	const { toast } = useToast();
	const [permissions, setPermissions] =
		useState<Record<string, Record<string, string[]>>>(initialPermissions);

	// Sync with parent data
	useEffect(() => {
		setPermissions(initialPermissions);
	}, [initialPermissions]);

	const hasPermission = (role: string, resource: string, action: string) =>
		permissions[role]?.[resource]?.includes(action) ?? false;

	const togglePermission = (role: string, resource: string, action: string) => {
		setPermissions((prev) => {
			const next = { ...prev };
			if (!next[role]) next[role] = {};
			if (!next[role][resource]) next[role][resource] = [];

			const granted = !next[role][resource].includes(action);
			next[role][resource] = granted
				? [...next[role][resource], action]
				: next[role][resource].filter((a) => a !== action);

			onPermissionChange?.(role, resource, action, granted);
			toast({
				title: granted ? "Permission Granted" : "Permission Revoked",
				description: `${action} on ${resource} for ${role}`,
			});
			return { ...next };
		});
	};

	const rolePermCount = (role: string) =>
		Object.values(permissions[role] ?? {}).reduce(
			(sum, acts) => sum + acts.length,
			0,
		);

	return (
		<NexusCard className="overflow-hidden">
			<div className="overflow-x-auto">
				<table className="w-full text-sm">
					<thead>
						<tr className="border-border bg-muted/50 border-b">
							<th className="text-muted-foreground bg-muted/50 sticky left-0 z-10 min-w-[140px] px-4 py-3 text-left font-medium">
								Resource / Action
							</th>
							{roles.map((role) => (
								<th
									key={role}
									colSpan={actions.length}
									className="text-foreground border-border border-l px-2 py-3 text-center font-medium"
								>
									<div className="flex items-center justify-center gap-2">
										{role}
										<Badge variant="secondary" className="text-[10px]">
											{rolePermCount(role)}
										</Badge>
									</div>
								</th>
							))}
						</tr>
						<tr className="border-border bg-muted/30 border-b">
							<th className="bg-muted/30 sticky left-0 z-10" />
							{roles.map((role) =>
								actions.map((action) => (
									<th
										key={`${role}-${action}`}
										className="text-muted-foreground border-border border-l px-2 py-2 text-center text-[11px] font-medium tracking-wider uppercase first:border-l-0"
									>
										{action.charAt(0)}
									</th>
								)),
							)}
						</tr>
					</thead>
					<tbody>
						{resources.map((resource) => (
							<tr
								key={resource}
								className="border-border hover:bg-muted/20 border-b transition-colors last:border-0"
							>
								<td className="text-foreground bg-background sticky left-0 z-10 px-4 py-3 font-medium">
									{resource}
								</td>
								{roles.map((role) =>
									actions.map((action) => (
										<td
											key={`${role}-${resource}-${action}`}
											className="border-border border-l px-2 py-3 text-center"
										>
											<Checkbox
												checked={hasPermission(role, resource, action)}
												onCheckedChange={() =>
													togglePermission(role, resource, action)
												}
												className={cn(
													"mx-auto",
													hasPermission(role, resource, action) &&
														"data-[state=checked]:bg-primary",
												)}
											/>
										</td>
									)),
								)}
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</NexusCard>
	);
}
