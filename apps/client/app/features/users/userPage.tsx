import { useState, useMemo } from "react";
import { z } from "zod";
import { PageHeader } from "@/components/layout/page-header";
import {  NexusButton  } from "@casbin/ui";
import {  NexusBadge  } from "@casbin/ui";
import { CrudTable, CrudFormDialog, DeleteDialog, type CrudColumnDef, type FieldDef } from "@/features/shared";
import { Plus } from "lucide-react";
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from "./userHooks";
import type { User } from "@/lib/api/schemas";
import {  Skeleton  } from "@casbin/ui";

/* ── Fallback mock data (used when API is not available) ── */
const mockUsers: (User & { id: string })[] = [
  { id: "1", name: "Alice Johnson", email: "alice@nexus.io", username: "alice", role: "Admin", status: "active", created_at: 1700000000 },
  { id: "2", name: "Bob Smith", email: "bob@nexus.io", username: "bob", role: "Editor", status: "active", created_at: 1700100000 },
  { id: "3", name: "Carol White", email: "carol@nexus.io", username: "carol", role: "Viewer", status: "suspended", created_at: 1700200000 },
  { id: "4", name: "David Brown", email: "david@nexus.io", username: "david", role: "Admin", status: "active", created_at: 1700300000 },
  { id: "5", name: "Eve Davis", email: "eve@nexus.io", username: "eve", role: "Editor", status: "active", created_at: 1700400000 },
  { id: "6", name: "Frank Miller", email: "frank@nexus.io", username: "frank", role: "Viewer", status: "active", created_at: 1700500000 },
  { id: "7", name: "Grace Lee", email: "grace@nexus.io", username: "grace", role: "Admin", status: "active", created_at: 1700600000 },
  { id: "8", name: "Henry Wilson", email: "henry@nexus.io", username: "henry", role: "Editor", status: "suspended", created_at: 1700700000 },
];

type UserRow = User & { id: string };

const columns: CrudColumnDef<UserRow>[] = [
  {
    id: "name", header: "Name", accessorKey: "name", sortable: true, minWidth: 200,
    cell: (row) => (
      <div>
        <p className="font-medium text-foreground">{row.name}</p>
        <p className="text-xs text-muted-foreground">{row.email}</p>
      </div>
    ),
  },
  { id: "username", header: "Username", accessorKey: "username", sortable: true },
  {
    id: "role", header: "Role", accessorKey: "role", sortable: true,
    filterable: true,
    filterOptions: [
      { label: "Admin", value: "Admin" },
      { label: "Editor", value: "Editor" },
      { label: "Viewer", value: "Viewer" },
    ],
  },
  {
    id: "status", header: "Status", accessorKey: "status",
    filterable: true,
    filterOptions: [{ label: "Active", value: "active" }, { label: "Suspended", value: "suspended" }],
    cell: (row) => (
      <NexusBadge variant={row.status === "active" ? "success" : "danger"}>{row.status}</NexusBadge>
    ),
  },
];

const createSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  email: z.string().trim().email("Invalid email").max(255),
  username: z.string().trim().min(2, "Min 2 characters").max(50),
  password: z.string().min(8, "Min 8 characters"),
  role: z.string().min(1, "Role is required"),
});

const editSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  email: z.string().trim().email("Invalid email").max(255),
  username: z.string().trim().min(2, "Min 2 characters").max(50),
  role: z.string().min(1, "Role is required"),
  status: z.string().min(1, "Status is required"),
});

const createFields: FieldDef[] = [
  { name: "name", label: "Full Name", type: "text", required: true, placeholder: "e.g. John Doe" },
  { name: "email", label: "Email", type: "email", required: true, placeholder: "john@example.com" },
  { name: "username", label: "Username", type: "text", required: true, placeholder: "johndoe" },
  { name: "password", label: "Password", type: "password", required: true, placeholder: "Min 8 characters" },
  {
    name: "role", label: "Role", type: "select", required: true,
    options: [{ label: "Admin", value: "Admin" }, { label: "Editor", value: "Editor" }, { label: "Viewer", value: "Viewer" }],
  },
];

const editFields: FieldDef[] = [
  { name: "name", label: "Full Name", type: "text", required: true },
  { name: "email", label: "Email", type: "email", required: true },
  { name: "username", label: "Username", type: "text", required: true },
  {
    name: "role", label: "Role", type: "select", required: true,
    options: [{ label: "Admin", value: "Admin" }, { label: "Editor", value: "Editor" }, { label: "Viewer", value: "Viewer" }],
  },
  {
    name: "status", label: "Status", type: "select", required: true,
    options: [{ label: "Active", value: "active" }, { label: "Suspended", value: "suspended" }],
  },
];

export default function UsersPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [editItem, setEditItem] = useState<UserRow | null>(null);
  const [deleteItem, setDeleteItem] = useState<UserRow | null>(null);

  const { data: usersResponse, isLoading, isError } = useUsers();
  const createUser = useCreateUser();
  const updateUser = useUpdateUser();
  const deleteUser = useDeleteUser();

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
            <Plus className="h-4 w-4" /> Add User
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
          bulkActions={[{
            label: "Delete Selected",
            onClick: (ids) => {
              ids.forEach(id => deleteUser.mutate(String(id)));
            },
          }]}
        />
      )}

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
            await updateUser.mutateAsync({ id: editItem.id, data: values as any });
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
