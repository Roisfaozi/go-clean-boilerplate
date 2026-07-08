import { useState, useMemo } from "react";
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
import {
	useEndpoints,
	useCreateEndpoint,
	useUpdateEndpoint,
	useDeleteEndpoint,
} from "./endpointHooks";
import { useResources } from "@/features/resources/resourceHooks";
import type { AccessRight, Endpoint } from "@/lib/api/types";

const _mockData: Endpoint[] = [
	{
		id: "1",
		name: "List Users",
		method: "GET",
		path: "/api/v1/users",
		resource_id: "1",
		resource_name: "Users",
		description: "Get paginated user list",
		auth_required: true,
		status: "active",
		created_at: 1700000000,
	},
	{
		id: "2",
		name: "Create User",
		method: "POST",
		path: "/api/v1/users",
		resource_id: "1",
		resource_name: "Users",
		description: "Create a new user",
		auth_required: true,
		status: "active",
		created_at: 1700100000,
	},
	{
		id: "3",
		name: "Update User",
		method: "PUT",
		path: "/api/v1/users/:id",
		resource_id: "1",
		resource_name: "Users",
		description: "Update user by ID",
		auth_required: true,
		status: "active",
		created_at: 1700200000,
	},
	{
		id: "4",
		name: "Delete User",
		method: "DELETE",
		path: "/api/v1/users/:id",
		resource_id: "1",
		resource_name: "Users",
		description: "Delete user by ID",
		auth_required: true,
		status: "active",
		created_at: 1700300000,
	},
	{
		id: "5",
		name: "List Projects",
		method: "GET",
		path: "/api/v1/projects",
		resource_id: "2",
		resource_name: "Projects",
		description: "Get all projects",
		auth_required: true,
		status: "active",
		created_at: 1700400000,
	},
	{
		id: "6",
		name: "Health Check",
		method: "GET",
		path: "/api/v1/health",
		resource_id: "6",
		resource_name: "System",
		description: "Public health check",
		auth_required: false,
		status: "active",
		created_at: 1700500000,
	},
];

const methodColor = (m: string) => {
	switch (m) {
		case "GET":
			return "info" as const;
		case "POST":
			return "success" as const;
		case "PUT":
		case "PATCH":
			return "warning" as const;
		case "DELETE":
			return "danger" as const;
		default:
			return "neutral" as const;
	}
};

const columns: CrudColumnDef<Endpoint>[] = [
	{
		id: "method",
		header: "Method",
		accessorKey: "method",
		minWidth: 100,
		filterable: true,
		filterOptions: [
			{ label: "GET", value: "GET" },
			{ label: "POST", value: "POST" },
			{ label: "PUT", value: "PUT" },
			{ label: "PATCH", value: "PATCH" },
			{ label: "DELETE", value: "DELETE" },
		],
		cell: (row) => (
			<NexusBadge variant={methodColor(row.method)}>{row.method}</NexusBadge>
		),
	},
	{
		id: "name",
		header: "Name",
		accessorKey: "name",
		sortable: true,
		minWidth: 160,
	},
	{
		id: "path",
		header: "Path",
		accessorKey: "path",
		minWidth: 220,
		cell: (row) => (
			<code className="bg-muted rounded px-2 py-1 font-mono text-xs">
				{row.path}
			</code>
		),
	},
	{
		id: "resource_name",
		header: "Resource",
		accessorKey: "resource_name",
		sortable: true,
	},
	{
		id: "auth_required",
		header: "Auth",
		accessorKey: "auth_required",
		cell: (row) => (
			<NexusBadge variant={row.auth_required ? "warning" : "neutral"}>
				{row.auth_required ? "Required" : "Public"}
			</NexusBadge>
		),
	},
];

const schema = z.object({
	name: z.string().trim().min(1, "Name is required").max(200),
	method: z.string().min(1, "Method is required"),
	path: z.string().trim().min(1, "Path is required").max(500),
	resource_id: z.string().min(1, "Resource is required"),
	description: z.string().max(500).optional(),
	auth_required: z.boolean().optional(),
});

const baseFields: FieldDef[] = [
	{
		name: "name",
		label: "Endpoint Name",
		type: "text",
		required: true,
		placeholder: "e.g. List Users",
	},
	{
		name: "method",
		label: "HTTP Method",
		type: "select",
		required: true,
		options: [
			{ label: "GET", value: "GET" },
			{ label: "POST", value: "POST" },
			{ label: "PUT", value: "PUT" },
			{ label: "PATCH", value: "PATCH" },
			{ label: "DELETE", value: "DELETE" },
		],
	},
	{
		name: "path",
		label: "Path",
		type: "text",
		required: true,
		placeholder: "/api/v1/users",
	},
	{
		name: "description",
		label: "Description",
		type: "textarea",
		placeholder: "Describe what this endpoint does…",
	},
	{ name: "auth_required", label: "Authentication Required", type: "switch" },
];

export default function EndpointsPage() {
	const [createOpen, setCreateOpen] = useState(false);
	const [editItem, setEditItem] = useState<Endpoint | null>(null);
	const [deleteItem, setDeleteItem] = useState<Endpoint | null>(null);

	const { data: response, isLoading, refetch } = useEndpoints();
	const { data: resourcesResponse } = useResources();
	const createEndpoint = useCreateEndpoint();
	const updateEndpoint = useUpdateEndpoint();
	const deleteEndpoint = useDeleteEndpoint();

	const endpoints: Endpoint[] = useMemo(() => {
		if (response?.data) return response.data as Endpoint[];
		return [];
	}, [response]);

	const resources: AccessRight[] = useMemo(() => {
		if (resourcesResponse?.data) return resourcesResponse.data as AccessRight[];
		return [];
	}, [resourcesResponse]);

	const fields: FieldDef[] = useMemo(() => {
		const resourceField: FieldDef = {
			name: "resource_id",
			label: "Resource",
			type: "select",
			required: true,
			options: resources.map((r) => ({ label: r.name, value: r.id })),
		};
		// Insert resource field after "path" (index 2)
		const result = [...baseFields];
		result.splice(3, 0, resourceField);
		return result;
	}, [resources]);

	return (
		<div className="space-y-6">
			<PageHeader
				title="Endpoints"
				description="Register and manage API endpoints."
				actions={
					<div className="flex gap-2">
						<NexusButton variant="outline" onClick={() => refetch()}>
							Refresh
						</NexusButton>
						<NexusButton onClick={() => setCreateOpen(true)}>
							<Plus className="mr-2 h-4 w-4" /> New Endpoint
						</NexusButton>
					</div>
				}
			/>

			<CrudTable
				columns={columns}
				data={endpoints}
				loading={isLoading}
				onEdit={setEditItem}
				onDelete={setDeleteItem}
			/>

			<CrudFormDialog
				open={createOpen}
				onOpenChange={setCreateOpen}
				title="Register Endpoint"
				description="Add a new API endpoint."
				fields={fields}
				schema={schema}
				loading={createEndpoint.isPending}
				onSubmit={(v) => {
					createEndpoint.mutate(v as any);
					setCreateOpen(false);
				}}
				submitLabel="Create Endpoint"
			/>
			<CrudFormDialog
				open={!!editItem}
				onOpenChange={(o) => !o && setEditItem(null)}
				title="Edit Endpoint"
				description="Update endpoint details."
				fields={fields}
				schema={schema}
				loading={updateEndpoint.isPending}
				initialValues={editItem || undefined}
				onSubmit={(v) => {
					if (editItem) {
						updateEndpoint.mutate({
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
				resourceName="Endpoint"
				itemName={deleteItem?.name}
				loading={deleteEndpoint.isPending}
				onConfirm={() => {
					if (deleteItem) {
						deleteEndpoint.mutate(String(deleteItem.id));
						setDeleteItem(null);
					}
				}}
			/>
		</div>
	);
}
