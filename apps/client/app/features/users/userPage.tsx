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
import { Plus, UserCog } from "lucide-react";
import {
	useUsers,
	useCreateUser,
	useUpdateUser,
	useDeleteUser,
} from "./userHooks";
import { RoleAssignmentModal } from "./RoleAssignmentModal";
import type { User } from "@/lib/api/schemas";
import { Skeleton } from "@casbin/ui";

type UserRow = User & { id: string; roles?: string[] };

/* ── Fallback mock data ── */
const mockUsers: UserRow[] = [
	{
		id: "1",
		name: "Alice Johnson",
		email: "alice@nexus.io",
		username: "alice",
		roles: ["Admin"],
		status: "active",
		created_at: 1700000000,
	},
	{
		id: "2",
		name: "Bob Smith",
		email: "bob@nexus.io",
		username: "bob",
		roles: ["Editor"],
		status: "active",
		created_at: 1700100000,
	},
];

const createSchema = z.object({
	name: z.string().trim().min(1, "Name is required"),
	email: z.string().trim().email("Invalid email"),
	username: z.string().trim().min(2, "Min 2 chars"),
	password: z.string().min(8, "Min 8 chars"),
});

const editSchema = z.object({
	name: z.string().trim().min(1, "Name is required"),
	email: z.string().trim().email("Invalid email"),
	username: z.string().trim().min(2, "Min 2 chars"),
	status: z.string().min(1, "Status is required"),
});

const createFields: FieldDef[] = [
	{ name: "name", label: "Full Name", type: "text", required: true },
	{ name: "email", label: "Email", type: "email", required: true },
	{ name: "username", label: "Username", type: "text", required: true },
	{ name: "password", label: "Password", type: "password", required: true },
];

const editFields: FieldDef[] = [
	{ name: "name", label: "Full Name", type: "text", required: true },
	{ name: "email", label: "Email", type: "email", required: true },
	{ name: "username", label: "Username", type: "text", required: true },
	{
		name: "status",
		label: "Status",
		type: "select",
		required: true,
		options: [
			{ label: "Active", value: "active" },
			{ label: "Suspended", value: "suspended" },
		],
	},
];

export default function UsersPage() {
	const [createOpen, setCreateOpen] = useState(false);
	const [editItem, setEditItem] = useState<UserRow | null>(null);
	const [deleteItem, setDeleteItem] = useState<UserRow | null>(null);
	const [roleManageItem, setRoleManageItem] = useState<UserRow | null>(null);

	const { data: usersResponse, isLoading, isError } = useUsers();
	const createUser = useCreateUser();
	const updateUser = useUpdateUser();
	const deleteUser = useDeleteUser();

	const columns: CrudColumnDef<UserRow>[] = useMemo(
		() => [
			{
				id: "name",
				header: "Name",
				accessorKey: "name",
				sortable: true,
				minWidth: 200,
				cell: (row) => (
					<div className="flex items-center gap-3">
						<div className="bg-primary/10 text-primary flex h-8 w-8 items-center justify-center rounded-full text-xs font-bold">
							{row.name.charAt(0)}
						</div>
						<div>
							<p className="text-foreground font-medium">{row.name}</p>
							<p className="text-muted-foreground text-xs">{row.email}</p>
						</div>
					</div>
				),
			},
			{
				id: "username",
				header: "Username",
				accessorKey: "username",
				sortable: true,
			},
			{
				id: "roles",
				header: "Roles",
				cell: (row) => (
					<div className="flex max-w-[200px] flex-wrap gap-1">
						{row.roles && row.roles.length > 0 ? (
							row.roles.map((r: string) => (
								<NexusBadge
									key={r}
									variant="neutral"
									className="px-1.5 py-0 text-[10px]"
								>
									{r}
								</NexusBadge>
							))
						) : (
							<span className="text-muted-foreground text-[10px] italic">
								No roles
							</span>
						)}
						<NexusButton
							variant="ghost"
							size="icon"
							className="ml-1 h-5 w-5 opacity-0 transition-opacity group-hover:opacity-100"
							onClick={() => setRoleManageItem(row)}
						>
							<UserCog className="text-primary h-3 w-3" />
						</NexusButton>
					</div>
				),
			},
			{
				id: "status",
				header: "Status",
				accessorKey: "status",
				filterable: true,
				filterOptions: [
					{ label: "Active", value: "active" },
					{ label: "Suspended", value: "suspended" },
				],
				cell: (row) => (
					<NexusBadge variant={row.status === "active" ? "success" : "danger"}>
						{row.status}
					</NexusBadge>
				),
			},
		],
		[],
	);

	// Use API data if available, fallback to mock
	const users: UserRow[] = useMemo(() => {
		if (usersResponse?.data) return usersResponse.data as UserRow[];
		if (isError) return mockUsers;
		return mockUsers;
	}, [usersResponse, isError]);

	return (
		<div className="space-y-6">
			<PageHeader
				title="Users"
				description="Manage user accounts and permissions."
				actions={
					<NexusButton onClick={() => setCreateOpen(true)}>
						<Plus className="mr-2 h-4 w-4" /> Add User
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
					data={users}
					selectable
					onEdit={setEditItem}
					onDelete={setDeleteItem}
					bulkActions={[
						{
							label: "Delete Selected",
							onClick: (ids) => {
								ids.forEach((id) => {
									deleteUser.mutate(String(id));
								});
							},
						},
					]}
				/>
			)}

			<RoleAssignmentModal
				open={!!roleManageItem}
				onOpenChange={(open) => !open && setRoleManageItem(null)}
				userId={roleManageItem?.id || ""}
				userName={roleManageItem?.name || ""}
				currentRoles={roleManageItem?.roles || []}
			/>

			<CrudFormDialog
				open={createOpen}
				onOpenChange={setCreateOpen}
				title="Create User"
				description="Add a new user to the system."
				fields={createFields}
				schema={createSchema}
				onSubmit={async (values) => {
					await createUser.mutateAsync(values as any);
					setCreateOpen(false);
				}}
				submitLabel="Create User"
			/>

			<CrudFormDialog
				open={!!editItem}
				onOpenChange={(open) => !open && setEditItem(null)}
				title="Edit User"
				description="Update user details."
				fields={editFields}
				schema={editSchema}
				initialValues={editItem || undefined}
				onSubmit={async (values) => {
					if (editItem) {
						await updateUser.mutateAsync({
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
				resourceName="User"
				itemName={deleteItem?.name}
				onConfirm={async () => {
					if (deleteItem) {
						await deleteUser.mutateAsync(String(deleteItem.id));
						setDeleteItem(null);
					}
				}}
			/>
		</div>
	);
}
