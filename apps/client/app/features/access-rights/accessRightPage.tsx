import { useState } from "react";
import { z } from "zod";
import { PageHeader } from "@/components/layout/page-header";
import { NexusButton } from "@casbin/ui";
import { NexusBadge } from "@casbin/ui";
import {
	CrudTable,
	CrudFormDialog,
	DeleteDialog,
	type CrudColumnDef,
	type FieldDef,
} from "@/features/shared";
import { Plus } from "lucide-react";
import type { AccessRight } from "@/lib/api/types";

const mockData: AccessRight[] = [
	{
		id: "1",
		name: "Read Users",
		resource: "users",
		action: "read",
		created_at: 1700000000,
	},
	{
		id: "2",
		name: "Write Users",
		resource: "users",
		action: "write",
		created_at: 1700100000,
	},
	{
		id: "3",
		name: "Delete Users",
		resource: "users",
		action: "delete",
		created_at: 1700200000,
	},
	{
		id: "4",
		name: "Read Projects",
		resource: "projects",
		action: "read",
		created_at: 1700300000,
	},
	{
		id: "5",
		name: "Manage Roles",
		resource: "roles",
		action: "manage",
		created_at: 1700400000,
	},
	{
		id: "6",
		name: "Read Organizations",
		resource: "organizations",
		action: "read",
		created_at: 1700500000,
	},
	{
		id: "7",
		name: "Write Projects",
		resource: "projects",
		action: "write",
		created_at: 1700600000,
	},
];

const actionColor = (a: string) =>
	a === "read"
		? ("info" as const)
		: a === "write"
			? ("warning" as const)
			: a === "delete"
				? ("danger" as const)
				: ("neutral" as const);

const columns: CrudColumnDef<AccessRight>[] = [
	{
		id: "name",
		header: "Name",
		accessorKey: "name",
		sortable: true,
		minWidth: 180,
	},
	{
		id: "resource",
		header: "Resource",
		accessorKey: "resource",
		sortable: true,
		filterable: true,
		filterOptions: [
			{ label: "Users", value: "users" },
			{ label: "Projects", value: "projects" },
			{ label: "Roles", value: "roles" },
			{ label: "Organizations", value: "organizations" },
		],
	},
	{
		id: "action",
		header: "Action",
		accessorKey: "action",
		filterable: true,
		filterOptions: [
			{ label: "Read", value: "read" },
			{ label: "Write", value: "write" },
			{ label: "Delete", value: "delete" },
			{ label: "Manage", value: "manage" },
		],
		cell: (row) => (
			<NexusBadge variant={actionColor(row.action)}>{row.action}</NexusBadge>
		),
	},
];

const schema = z.object({
	name: z.string().trim().min(1, "Name is required").max(100),
	resource: z.string().min(1, "Resource is required"),
	action: z.string().min(1, "Action is required"),
});

const fields: FieldDef[] = [
	{
		name: "name",
		label: "Name",
		type: "text",
		required: true,
		placeholder: "e.g. Read Users",
	},
	{
		name: "resource",
		label: "Resource",
		type: "select",
		required: true,
		options: [
			{ label: "Users", value: "users" },
			{ label: "Projects", value: "projects" },
			{ label: "Roles", value: "roles" },
			{ label: "Organizations", value: "organizations" },
		],
	},
	{
		name: "action",
		label: "Action",
		type: "select",
		required: true,
		options: [
			{ label: "Read", value: "read" },
			{ label: "Write", value: "write" },
			{ label: "Delete", value: "delete" },
			{ label: "Manage", value: "manage" },
		],
	},
];

export default function AccessRightsPage() {
	const [createOpen, setCreateOpen] = useState(false);
	const [editItem, setEditItem] = useState<AccessRight | null>(null);
	const [deleteItem, setDeleteItem] = useState<AccessRight | null>(null);

	return (
		<div className="space-y-6">
			<PageHeader
				title="Access Rights"
				description="Manage resource-level access rights."
				actions={
					<NexusButton onClick={() => setCreateOpen(true)}>
						<Plus className="h-4 w-4" /> New Access Right
					</NexusButton>
				}
			/>
			<CrudTable
				columns={columns}
				data={mockData}
				onEdit={setEditItem}
				onDelete={setDeleteItem}
			/>
			<CrudFormDialog
				open={createOpen}
				onOpenChange={setCreateOpen}
				title="Create Access Right"
				fields={fields}
				schema={schema}
				onSubmit={async (v) => console.log("create", v)}
				submitLabel="Create"
			/>
			<CrudFormDialog
				open={!!editItem}
				onOpenChange={(o) => !o && setEditItem(null)}
				title="Edit Access Right"
				fields={fields}
				schema={schema}
				initialValues={editItem || undefined}
				onSubmit={async (v) => console.log("update", editItem?.id, v)}
				submitLabel="Save Changes"
			/>
			<DeleteDialog
				open={!!deleteItem}
				onOpenChange={(o) => !o && setDeleteItem(null)}
				resourceName="Access Right"
				itemName={deleteItem?.name}
				onConfirm={async () => console.log("delete", deleteItem?.id)}
			/>
		</div>
	);
}
