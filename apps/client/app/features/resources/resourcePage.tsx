import { PageHeader } from "@/components/layout/page-header";
import {
	CrudFormDialog,
	CrudTable,
	DeleteDialog,
	type CrudColumnDef,
	type FieldDef,
} from "@/features/shared";
import type { AccessRight } from "@/lib/api/types";
import { NexusButton } from "@casbin/ui";
import { Plus } from "lucide-react";
import { useMemo, useState } from "react";
import { z } from "zod";
import {
	useCreateResource,
	useDeleteResource,
	useResources,
	useUpdateResource,
} from "./resourceHooks";

const columns: CrudColumnDef<AccessRight>[] = [
	{
		id: "name",
		header: "Name",
		accessorKey: "name",
		sortable: true,
		minWidth: 180,
	},
	{
		id: "description",
		header: "Description",
		accessorKey: "description",
		minWidth: 240,
	},
];

const createSchema = z.object({
	name: z.string().trim().min(3, "Name must be at least 3 characters").max(100),
	description: z.string().trim().max(255).optional(),
});

const editSchema = createSchema;

const createFields: FieldDef[] = [
	{
		name: "name",
		label: "Access Right Name",
		type: "text",
		required: true,
		placeholder: "e.g. manage_users",
	},
	{
		name: "description",
		label: "Description",
		type: "textarea",
		placeholder: "What this access right grants",
	},
];

const editFields: FieldDef[] = [...createFields];

export default function ResourcesPage() {
	const [createOpen, setCreateOpen] = useState(false);
	const [editItem, setEditItem] = useState<AccessRight | null>(null);
	const [deleteItem, setDeleteItem] = useState<AccessRight | null>(null);

	const { data: response, isLoading, refetch } = useResources();
	const createResource = useCreateResource();
	const updateResource = useUpdateResource();
	const deleteResource = useDeleteResource();

	const resources: AccessRight[] = useMemo(() => {
		if (response?.data) return response.data as AccessRight[];
		return [];
	}, [response]);

	return (
		<div className="space-y-6">
			<PageHeader
				title="Access Rights"
				description="Register and manage access rights exposed by the backend."
				actions={
					<div className="flex gap-2">
						<NexusButton variant="outline" onClick={() => refetch()}>
							Refresh
						</NexusButton>
						<NexusButton onClick={() => setCreateOpen(true)}>
							<Plus className="mr-2 h-4 w-4" /> New Access Right
						</NexusButton>
					</div>
				}
			/>

			<CrudTable
				columns={columns}
				data={resources}
				loading={isLoading}
				onEdit={setEditItem}
				onDelete={setDeleteItem}
			/>

			<CrudFormDialog
				open={createOpen}
				onOpenChange={setCreateOpen}
				title="Register Access Right"
				description="Add a new access right mapping."
				fields={createFields}
				schema={createSchema}
				loading={createResource.isPending}
				onSubmit={(v) => {
					createResource.mutate(v as any);
					setCreateOpen(false);
				}}
				submitLabel="Create Access Right"
			/>
			<CrudFormDialog
				open={!!editItem}
				onOpenChange={(o) => !o && setEditItem(null)}
				title="Edit Access Right"
				description="Update access right details."
				fields={editFields}
				schema={editSchema}
				loading={updateResource.isPending}
				initialValues={editItem || undefined}
				onSubmit={(v) => {
					if (editItem) {
						updateResource.mutate({
							id: editItem.id,
							data: v as any,
						});
						setEditItem(null);
					}
				}}
				submitLabel="Save Changes"
			/>
			<DeleteDialog
				open={!!deleteItem}
				onOpenChange={(o) => !o && setDeleteItem(null)}
				resourceName="Access Right"
				itemName={deleteItem?.name}
				loading={deleteResource.isPending}
				onConfirm={() => {
					if (deleteItem) {
						deleteResource.mutate(String(deleteItem.id));
						setDeleteItem(null);
					}
				}}
			/>
		</div>
	);
}
