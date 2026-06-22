import { PageHeader } from "@/components/layout/page-header";
import {
	CrudFormDialog,
	CrudTable,
	DeleteDialog,
	type CrudColumnDef,
} from "@/features/shared";
import type { Permission, Role } from "@/lib/api/types";
import {
	NexusBadge,
	NexusButton,
	NexusCard,
	Skeleton,
	Switch,
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
	Tabs,
	TabsContent,
	TabsList,
	TabsTrigger,
} from "@casbin/ui";
import {
	ArrowRight,
	Box,
	GitBranch,
	Key,
	Plus,
	RefreshCw,
	Shield,
	Trash2,
} from "lucide-react";
import { useMemo, useState } from "react";
import { z } from "zod";
import { useRoles } from "../roles/roleHooks";
import {
	useAddInheritance,
	useCreatePermission,
	useDeletePermission,
	useInheritanceTree,
	usePermissions,
	useRemoveInheritance,
	useResourceAggregation,
	useRoleAccessRights,
	useToggleAccessRight,
	useUpdatePermission,
} from "./permissionHooks";
import { RoleInheritanceTree, type RoleNode } from "./role-inheritance-tree";

function findNodeById(nodes: RoleNode[], id: string): RoleNode | null {
	for (const node of nodes) {
		if (node.id === id) return node;
		if (node.children) {
			const found = findNodeById(node.children, id);
			if (found) return found;
		}
	}
	return null;
}

// --- Matrix Components ---

function MatrixCell({
	role,
	resourceId,
	domain,
}: {
	role: string;
	resourceId: string;
	domain?: string;
}) {
	const { data: roleData, isLoading } = useRoleAccessRights(role, domain);
	const toggle = useToggleAccessRight();

	const status = useMemo(() => {
		const data = Array.isArray(roleData) ? roleData : [];
		const item = data.find((r: any) => r.id === resourceId);
		return {
			assigned: item?.is_assigned ?? false,
			partial: item?.is_partial ?? false,
		};
	}, [roleData, resourceId]);

	if (isLoading) return <Skeleton className="mx-auto h-5 w-10" />;

	return (
		<div className="flex flex-col items-center gap-1">
			<Switch
				checked={status.assigned}
				onCheckedChange={(checked) =>
					toggle.mutate({
						role,
						access_right_id: resourceId,
						granted: checked,
						domain,
					})
				}
				disabled={toggle.isPending}
			/>
			{status.partial && !status.assigned && (
				<span className="text-[10px] font-medium text-amber-500">Partial</span>
			)}
		</div>
	);
}

function PermissionMatrix({
	roles,
	resources,
}: {
	roles: Role[];
	resources: any[];
}) {
	return (
		<NexusCard className="shadow-premium overflow-hidden border-none bg-white/50 backdrop-blur-sm">
			<div className="overflow-x-auto">
				<Table>
					<TableHeader className="bg-muted/50">
						<TableRow>
							<TableHead className="text-foreground w-[250px] py-6 pl-8 font-bold">
								<div className="flex items-center gap-2">
									<Box className="text-primary h-4 w-4" />
									Resources / Modules
								</div>
							</TableHead>
							{roles.map((role) => (
								<TableHead
									key={role.id}
									className="text-foreground min-w-[120px] text-center font-bold"
								>
									<div className="flex flex-col items-center gap-1">
										<Shield className="text-primary/70 h-4 w-4" />
										{role.name}
									</div>
								</TableHead>
							))}
						</TableRow>
					</TableHeader>
					<TableBody>
						{resources.map((res) => (
							<TableRow
								key={res.id}
								className="hover:bg-primary/5 group transition-colors"
							>
								<TableCell className="py-4 pl-8 font-medium">
									<div className="flex flex-col">
										<span className="text-foreground group-hover:text-primary transition-colors">
											{res.name}
										</span>
										<span className="text-muted-foreground text-xs font-normal">
											{res.endpoint_count} Endpoints
										</span>
									</div>
								</TableCell>
								{roles.map((role) => (
									<TableCell
										key={`${res.id}-${role.id}`}
										className="text-center"
									>
										<MatrixCell role={role.name} resourceId={res.id} />
									</TableCell>
								))}
							</TableRow>
						))}
					</TableBody>
				</Table>
			</div>
		</NexusCard>
	);
}

function RoleInheritanceDetail({
	role,
	allRoles,
	onInheritanceChange,
}: {
	role: RoleNode;
	allRoles: Role[];
	onInheritanceChange: () => void;
}) {
	const addInheritance = useAddInheritance();
	const removeInheritance = useRemoveInheritance();

	const handleAddParent = async (parentRole: string) => {
		await addInheritance.mutateAsync({
			child_role: role.name,
			parent_role: parentRole,
		});
		onInheritanceChange();
	};

	const handleRemoveParent = async (parentRole: string) => {
		await removeInheritance.mutateAsync({
			child_role: role.name,
			parent_role: parentRole,
		});
		onInheritanceChange();
	};

	// Filter out self and current parents from available roles to add
	const availableRoles = allRoles.filter(
		(r) => r.name !== role.name && !(role.parents || []).includes(r.name),
	);

	return (
		<div className="space-y-6">
			<NexusCard className="shadow-premium overflow-hidden border-none">
				<div className="bg-primary/5 border-primary/10 border-b p-6">
					<div className="mb-4 flex items-center justify-between">
						<div className="flex items-center gap-3">
							<div className="bg-primary/10 rounded-lg p-2">
								<Shield className="text-primary h-6 w-6" />
							</div>
							<div>
								<h3 className="text-foreground text-lg font-bold">
									{role.name}
								</h3>
								<p className="text-muted-foreground text-sm">
									Role Management & Inheritance
								</p>
							</div>
						</div>
						<NexusBadge variant="neutral" className="px-3 py-1">
							ID: {role.id.substring(0, 8)}
						</NexusBadge>
					</div>
					{role.description && (
						<p className="text-muted-foreground rounded-md border border-white/20 bg-white/50 p-3 text-sm">
							{role.description}
						</p>
					)}
				</div>

				<div className="space-y-8 p-6">
					{/* Parents Management */}
					<section>
						<h4 className="text-foreground mb-4 flex items-center gap-2 font-semibold">
							<GitBranch className="text-primary h-4 w-4" />
							Inherits From (Parents)
						</h4>
						<div className="mb-4 flex flex-wrap gap-2">
							{(role.parents || []).length === 0 ? (
								<p className="text-muted-foreground text-xs italic">
									No parent roles assigned.
								</p>
							) : (
								role.parents?.map((parent) => (
									<NexusBadge
										key={parent}
										variant="info"
										className="flex items-center gap-1 px-2 py-1 pl-3"
									>
										{parent}
										<NexusButton
											variant="ghost"
											size="icon"
											className="ml-1 h-4 w-4 rounded-full p-0 hover:bg-white/20"
											onClick={() => handleRemoveParent(parent)}
											disabled={removeInheritance.isPending}
										>
											<Trash2 className="h-2.5 w-2.5" />
										</NexusButton>
									</NexusBadge>
								))
							)}
						</div>

						{availableRoles.length > 0 && (
							<div className="flex items-center gap-2 border-t border-dashed pt-2">
								<span className="text-muted-foreground text-xs font-medium">
									Add Parent:
								</span>
								<div className="flex flex-wrap gap-1.5">
									{availableRoles.slice(0, 5).map((r) => (
										<NexusButton
											key={r.id}
											variant="outline"
											size="sm"
											className="h-7 px-2 text-[10px]"
											onClick={() => handleAddParent(r.name)}
											disabled={addInheritance.isPending}
										>
											<Plus className="mr-1 h-3 w-3" /> {r.name}
										</NexusButton>
									))}
								</div>
							</div>
						)}
					</section>

					{/* Permissions Breakdown */}
					<section>
						<h4 className="text-foreground mb-4 flex items-center gap-2 font-semibold">
							<Key className="text-primary h-4 w-4" />
							Effective Permissions
						</h4>
						<div className="grid grid-cols-1 gap-3 md:grid-cols-2">
							{(role.effective_permissions || []).map((perm, idx) => {
								const isOwn = (role.own_permissions || []).some(
									(p) => p[0] === perm[0] && p[1] === perm[1],
								);
								return (
									<div
										key={idx}
										className="bg-muted/5 group hover:border-primary/30 flex items-center justify-between rounded-lg border p-3 transition-colors"
									>
										<div className="flex flex-col">
											<span className="text-foreground text-sm font-medium">
												{perm[0]}
											</span>
											<div className="mt-1 flex items-center gap-1.5">
												<span className="text-muted-foreground font-mono text-[10px]">
													{perm[1]}
												</span>
												<ArrowRight className="text-muted-foreground/30 h-2 w-2" />
												<span className="text-primary text-[10px] font-bold tracking-wider uppercase">
													{perm[2]}
												</span>
											</div>
										</div>
										<NexusBadge
											variant={isOwn ? "success" : "info"}
											className="px-1.5 py-0 text-[9px]"
										>
											{isOwn ? "Direct" : "Inherited"}
										</NexusBadge>
									</div>
								);
							})}
							{(role.effective_permissions || []).length === 0 && (
								<div className="bg-muted/5 col-span-2 rounded-xl border border-dashed py-8 text-center">
									<Shield className="text-muted/20 mx-auto mb-2 h-8 w-8" />
									<p className="text-muted-foreground text-xs italic">
										No permissions granted.
									</p>
								</div>
							)}
						</div>
					</section>
				</div>
			</NexusCard>
		</div>
	);
}

// --- Main Page ---

const columns: CrudColumnDef<Permission>[] = [
	{
		id: "role_name",
		header: "Role",
		accessorKey: "role_name",
		sortable: true,
		filterable: true,
		filterOptions: [
			{ label: "Admin", value: "Admin" },
			{ label: "Editor", value: "Editor" },
			{ label: "Viewer", value: "Viewer" },
		],
	},
	{
		id: "access_right_name",
		header: "Access Right",
		accessorKey: "access_right_name",
		sortable: true,
	},
	{
		id: "granted",
		header: "Status",
		accessorKey: "granted",
		cell: (row) => (
			<NexusBadge variant={row.granted ? "success" : "danger"}>
				{row.granted ? "Granted" : "Denied"}
			</NexusBadge>
		),
	},
];

export default function PermissionsPage() {
	const [activeTab, setActiveTab] = useState("matrix");
	const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null);
	const [createOpen, setCreateOpen] = useState(false);
	const [_editItem, setEditItem] = useState<Permission | null>(null);
	const [deleteItem, setDeleteItem] = useState<Permission | null>(null);

	// Queries
	const {
		data: response,
		isLoading: permissionsLoading,
		refetch: refetchPermissions,
	} = usePermissions();
	const { data: rolesResponse, isLoading: rolesLoading } = useRoles();
	const { data: resourcesResponse, isLoading: resourcesLoading } =
		useResourceAggregation();
	const { data: inheritanceResponse, isLoading: inheritanceLoading } =
		useInheritanceTree();

	// Mutations
	const createPermission = useCreatePermission();
	const _updatePermission = useUpdatePermission();
	const deletePermission = useDeletePermission();

	const permissions: Permission[] = useMemo(() => {
		return (response?.data || []) as Permission[];
	}, [response]);

	const roles = useMemo(
		() => (Array.isArray(rolesResponse) ? rolesResponse : []) as Role[],
		[rolesResponse],
	);
	const resources = useMemo(
		() => (resourcesResponse?.resources || []) as any[],
		[resourcesResponse],
	);
	const inheritanceTree = useMemo(
		() => (inheritanceResponse?.roles || []) as any[],
		[inheritanceResponse],
	);

	const selectedItem = useMemo(() => {
		if (!selectedRoleId) return null;
		return findNodeById(inheritanceTree, selectedRoleId);
	}, [inheritanceTree, selectedRoleId]);

	const isLoading =
		permissionsLoading ||
		rolesLoading ||
		resourcesLoading ||
		inheritanceLoading;

	return (
		<div className="space-y-6">
			<PageHeader
				title="Permissions Management"
				description="Configure access rights and security policies for your roles."
				actions={
					<div className="flex gap-2">
						<NexusButton
							variant="outline"
							size="sm"
							onClick={() => refetchPermissions()}
						>
							<RefreshCw className="mr-2 h-4 w-4" /> Refresh
						</NexusButton>
						<NexusButton size="sm" onClick={() => setCreateOpen(true)}>
							<Shield className="mr-2 h-4 w-4" /> Add Mapping
						</NexusButton>
					</div>
				}
			/>

			<Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
				<div className="mb-4 flex items-center justify-between">
					<TabsList className="bg-muted/50 p-1">
						<TabsTrigger
							value="matrix"
							className="data-[state=active]:bg-white data-[state=active]:shadow-sm"
						>
							<Key className="text-primary mr-2 h-4 w-4" /> Matrix View
						</TabsTrigger>
						<TabsTrigger
							value="inheritance"
							className="data-[state=active]:bg-white data-[state=active]:shadow-sm"
						>
							<GitBranch className="text-primary mr-2 h-4 w-4" /> Role
							Inheritance
						</TabsTrigger>
						<TabsTrigger
							value="list"
							className="data-[state=active]:bg-white data-[state=active]:shadow-sm"
						>
							<Shield className="text-primary mr-2 h-4 w-4" /> List View
						</TabsTrigger>
					</TabsList>
				</div>

				<TabsContent
					value="matrix"
					className="mt-0 outline-none focus-visible:ring-0"
				>
					{isLoading ? (
						<div className="grid grid-cols-1 gap-4">
							<Skeleton className="h-[400px] w-full rounded-xl" />
						</div>
					) : (
						<PermissionMatrix roles={roles} resources={resources} />
					)}
				</TabsContent>

				<TabsContent
					value="inheritance"
					className="mt-0 outline-none focus-visible:ring-0"
				>
					<div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
						<div className="lg:col-span-5">
							{isLoading ? (
								<Skeleton className="h-[500px] w-full rounded-xl" />
							) : (
								<RoleInheritanceTree
									tree={inheritanceTree}
									onSelect={(role) => setSelectedRoleId(role.id)}
									activeId={selectedRoleId || undefined}
								/>
							)}
						</div>
						<div className="lg:col-span-7">
							{selectedItem ? (
								<RoleInheritanceDetail
									role={selectedItem}
									allRoles={roles}
									onInheritanceChange={() => {
										// Refetch inheritance tree
										inheritanceResponse && refetchPermissions();
									}}
								/>
							) : (
								<NexusCard className="bg-muted/10 flex h-full min-h-[400px] flex-col items-center justify-center border-dashed p-12 text-center">
									<div className="bg-primary/10 mb-4 flex h-16 w-16 items-center justify-center rounded-full">
										<GitBranch className="text-primary h-8 w-8" />
									</div>
									<h4 className="text-foreground font-semibold">
										Select a Role
									</h4>
									<p className="text-muted-foreground mt-2 max-w-xs text-sm">
										Select a role from the tree to view its inheritance details,
										effective permissions, and manage its parent roles.
									</p>
								</NexusCard>
							)}
						</div>
					</div>
				</TabsContent>

				<TabsContent
					value="list"
					className="mt-0 outline-none focus-visible:ring-0"
				>
					<CrudTable
						columns={columns}
						data={permissions}
						loading={isLoading}
						onEdit={setEditItem}
						onDelete={setDeleteItem}
					/>
				</TabsContent>
			</Tabs>

			{/* CRUD Dialogs (keep for manual specific overrides if needed) */}
			<CrudFormDialog
				open={createOpen}
				onOpenChange={setCreateOpen}
				title="Manual Permission Mapping"
				description="Directly assign a specific access right to a role."
				fields={[
					{
						name: "role_id",
						label: "Role",
						type: "select",
						required: true,
						options: roles.map((r) => ({ label: r.name, value: r.id })),
					},
					{
						name: "access_right_id",
						label: "Access Right",
						type: "select",
						required: true,
						options: resources.map((r) => ({ label: r.name, value: r.id })),
					},
					{ name: "granted", label: "Granted", type: "switch" },
				]}
				schema={z.object({
					role_id: z.string().min(1),
					access_right_id: z.string().min(1),
					granted: z.boolean(),
				})}
				loading={createPermission.isPending}
				onSubmit={async (v) => {
					await createPermission.mutateAsync(v as any);
					setCreateOpen(false);
				}}
				submitLabel="Assign"
			/>

			<DeleteDialog
				open={!!deleteItem}
				onOpenChange={(o) => !o && setDeleteItem(null)}
				resourceName="Permission Mapping"
				itemName={`${deleteItem?.role_name} → ${deleteItem?.access_right_name}`}
				loading={deletePermission.isPending}
				onConfirm={async () => {
					if (deleteItem) {
						await deletePermission.mutateAsync(String(deleteItem.id));
						setDeleteItem(null);
					}
				}}
			/>
		</div>
	);
}
