import { useState, useMemo } from "react";
import { z } from "zod";
import { PageHeader } from "@/components/layout/page-header";
import { NexusButton } from "@casbin/ui";
import {
	CrudTable,
	CrudFormDialog,
	DeleteDialog,
	type CrudColumnDef,
	type FieldDef,
} from "@/features/shared";
import { Plus } from "lucide-react";
import {
	useRoles,
	useCreateRole,
	useUpdateRole,
	useDeleteRole,
} from "./roleHooks";
import type { Role } from "@/lib/api/schemas";
import { Skeleton } from "@casbin/ui";

const mockRoles: (Role & { id: string })[] = [
	{
		id: "1",
		name: "Admin",
		description: "Full access to all resources",
		created_at: 1700000000,
	},
	{
		id: "2",
		name: "Editor",
		description: "Can create and edit content",
		created_at: 1700100000,
	},
	{
		id: "3",
		name: "Viewer",
		description: "Read-only access",
		created_at: 1700200000,
	},
	{
		id: "4",
		name: "Moderator",
		description: "Can moderate user content",
		created_at: 1700300000,
	},
	{
		id: "5",
		name: "Manager",
		description: "Team management capabilities",
		created_at: 1700400000,
	},
];

type RoleRow = Role & { id: string };

const columns: CrudColumnDef<RoleRow>[] = [
	{
		id: "name",
		header: "Role Name",
		accessorKey: "name",
		sortable: true,
		minWidth: 160,
	},
	{
		id: "description",
		header: "Description",
		accessorKey: "description",
		minWidth: 250,
	},
];

const schema = z.object({
	name: z.string().trim().min(1, "Name is required").max(50),
	description: z.string().max(500).optional(),
});

const fields: FieldDef[] = [
	{
		name: "name",
		label: "Role Name",
		type: "text",
		required: true,
		placeholder: "e.g. Manager",
	},
	{
		name: "description",
		label: "Description",
		type: "textarea",
		placeholder: "Describe the role's permissions…",
	},
];

export default function RolesPage() {
	const [createOpen, setCreateOpen] = useState(false);
	const [editItem, setEditItem] = useState<RoleRow | null>(null);
	const [deleteItem, setDeleteItem] = useState<RoleRow | null>(null);

	const { data: rolesResponse, isLoading, isError } = useRoles();
	const createRole = useCreateRole();
	const updateRole = useUpdateRole();
	const deleteRole = useDeleteRole();

	const roles: RoleRow[] = useMemo(() => {
		if (rolesResponse?.data) return rolesResponse.data as RoleRow[];
		if (isError) return mockRoles;
		return mockRoles;
	}, [rolesResponse, isError]);

	return (
		<div className="space-y-6">
			<PageHeader
				title="Roles"
				description="Define and manage access roles."
				actions={
					<NexusButton onClick={() => setCreateOpen(true)}>
						<Plus className="h-4 w-4" /> Create Role
					</NexusButton>
				}
			/>

			{isLoading ? (
				<div className="space-y-3">
					{Array.from({ length: 5 }).map((_, i) => (
						<Skeleton key={i} className="h-14 w-full rounded-lg" />
					))}
				</div>
			) : (
				<CrudTable
					columns={columns}
					data={roles}
					onEdit={setEditItem}
					onDelete={setDeleteItem}
				/>
			)}

			<CrudFormDialog
				open={createOpen}
				onOpenChange={setCreateOpen}
				title="Create Role"
				description="Add a new role to the system."
				fields={fields}
				schema={schema}
				onSubmit={async (values) => {
					await createRole.mutateAsync(values as any);
					setCreateOpen(false);
				}}
				submitLabel="Create Role"
			/>

			<CrudFormDialog
				open={!!editItem}
				onOpenChange={(open) => !open && setEditItem(null)}
				title="Edit Role"
				description="Update role details."
				fields={fields}
				schema={schema}
				initialValues={editItem || undefined}
				onSubmit={async (values) => {
					if (editItem) {
						await updateRole.mutateAsync({
							id: editItem.id,
							data: values as any,
						});
						setEditItem(null);
					}
				}}
				submitLabel="Save Changes"
			/>

			<DeleteDialog
				open={!!deleteItem}
				onOpenChange={(open) => !open && setDeleteItem(null)}
				resourceName="Role"
				itemName={deleteItem?.name}
				onConfirm={async () => {
					if (deleteItem) {
						await deleteRole.mutateAsync(String(deleteItem.id));
						setDeleteItem(null);
					}
				}}
			/>
		</div>
	);
}
