import { useEffect, useMemo, useState } from "react";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogDescription,
	DialogFooter,
} from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { Checkbox } from "@casbin/ui";
import { Skeleton } from "@casbin/ui";
import { Badge } from "@casbin/ui";
import { Shield } from "lucide-react";
import { useRoles } from "../roles/roleHooks";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { permissionService } from "../permissions/permissionService";
import { toast } from "@casbin/ui";

interface RoleAssignmentModalProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	userId: string;
	userName: string;
	currentRoles: string[];
}

export function RoleAssignmentModal({
	open,
	onOpenChange,
	userId,
	userName,
	currentRoles,
}: RoleAssignmentModalProps) {
	const qc = useQueryClient();
	const { data: rolesResponse, isLoading: rolesLoading } = useRoles();
	const roles = useMemo(() => rolesResponse?.data || [], [rolesResponse]);

	// Track selected roles locally for immediate UI feedback
	const [selectedRoles, setSelectedRoles] = useState<string[]>(currentRoles);

	// Re-sync when currentRoles changes
	useEffect(() => {
		setSelectedRoles(currentRoles);
	}, [currentRoles.join(",")]);

	const assignMutation = useMutation({
		mutationFn: (roleName: string) =>
			permissionService.addInheritance({
				child_role: userId,
				parent_role: roleName,
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ["users"] });
			toast.success("Role assigned");
		},
		onError: () => toast.error("Failed to assign role"),
	});

	const revokeMutation = useMutation({
		mutationFn: (roleName: string) =>
			permissionService.removeInheritance({
				child_role: userId,
				parent_role: roleName,
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ["users"] });
			toast.success("Role revoked");
		},
		onError: () => toast.error("Failed to revoke role"),
	});

	const handleToggle = async (roleName: string, active: boolean) => {
		if (active) {
			setSelectedRoles((p) => [...p, roleName]);
			assignMutation.mutate(roleName, {
				onError: () => {
					setSelectedRoles((p) => p.filter((r) => r !== roleName));
				},
			});
		} else {
			setSelectedRoles((p) => p.filter((r) => r !== roleName));
			revokeMutation.mutate(roleName, {
				onError: () => {
					setSelectedRoles((p) => [...p, roleName]);
				},
			});
		}
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle className="flex items-center gap-2">
						<Shield className="text-primary h-5 w-5" />
						Manage Roles
					</DialogTitle>
					<DialogDescription>
						Assign or revoke security roles for <b>{userName}</b>.
					</DialogDescription>
				</DialogHeader>

				<div className="space-y-4 py-4">
					<div className="mb-2 flex flex-wrap gap-1.5">
						{selectedRoles.length === 0 ? (
							<span className="text-muted-foreground text-xs italic">
								No roles assigned
							</span>
						) : (
							selectedRoles.map((r) => (
								<Badge
									key={r}
									variant="secondary"
									className="px-2 py-0.5 text-[10px] font-bold"
								>
									{r}
								</Badge>
							))
						)}
					</div>

					<div className="max-h-[300px] divide-y overflow-y-auto rounded-lg border">
						{rolesLoading
							? Array.from({ length: 4 }).map((_, i) => (
									<div key={i} className="p-3">
										<Skeleton className="h-5 w-full" />
									</div>
								))
							: roles.map((role) => (
									<div
										key={role.id}
										className="hover:bg-muted/30 flex items-center justify-between p-3 transition-colors"
									>
										<div className="flex flex-col">
											<span className="text-sm font-medium">{role.name}</span>
											{role.description && (
												<span className="text-muted-foreground line-clamp-1 text-[10px]">
													{role.description}
												</span>
											)}
										</div>
										<Checkbox
											checked={selectedRoles.includes(role.name)}
											onCheckedChange={(checked) =>
												handleToggle(role.name, !!checked)
											}
										/>
									</div>
								))}
					</div>
				</div>

				<DialogFooter>
					<NexusButton onClick={() => onOpenChange(false)}>Close</NexusButton>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
