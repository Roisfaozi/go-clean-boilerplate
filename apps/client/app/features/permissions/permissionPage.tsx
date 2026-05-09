import { useState, useMemo } from "react";
import { z } from "zod";
import { PageHeader } from "@/components/layout/page-header";
import {  NexusButton  } from "@casbin/ui";
import {  NexusBadge  } from "@casbin/ui";
import { CrudTable, CrudFormDialog, DeleteDialog, type CrudColumnDef, type FieldDef } from "@/features/shared";
import { Plus } from "lucide-react";
import { usePermissions, useCreatePermission, useUpdatePermission, useDeletePermission } from "./permissionHooks";
import type { Permission } from "@/lib/api/types";
import {  Skeleton  } from "@casbin/ui";

const mockData: Permission[] = [
  { id: "1", role_id: "1", access_right_id: "1", granted: true, role_name: "Admin", access_right_name: "Read Users", created_at: 1700000000 },
  { id: "2", role_id: "1", access_right_id: "2", granted: true, role_name: "Admin", access_right_name: "Write Users", created_at: 1700100000 },
  { id: "3", role_id: "1", access_right_id: "3", granted: true, role_name: "Admin", access_right_name: "Delete Users", created_at: 1700200000 },
  { id: "4", role_id: "2", access_right_id: "1", granted: true, role_name: "Editor", access_right_name: "Read Users", created_at: 1700300000 },
  { id: "5", role_id: "2", access_right_id: "2", granted: true, role_name: "Editor", access_right_name: "Write Users", created_at: 1700400000 },
  { id: "6", role_id: "2", access_right_id: "3", granted: false, role_name: "Editor", access_right_name: "Delete Users", created_at: 1700500000 },
  { id: "7", role_id: "3", access_right_id: "1", granted: true, role_name: "Viewer", access_right_name: "Read Users", created_at: 1700600000 },
  { id: "8", role_id: "3", access_right_id: "4", granted: true, role_name: "Viewer", access_right_name: "Read Projects", created_at: 1700700000 },
];

const columns: CrudColumnDef<Permission>[] = [
  {
    id: "role_name", header: "Role", accessorKey: "role_name", sortable: true,
    filterable: true,
    filterOptions: [{ label: "Admin", value: "Admin" }, { label: "Editor", value: "Editor" }, { label: "Viewer", value: "Viewer" }],
  },
  { id: "access_right_name", header: "Access Right", accessorKey: "access_right_name", sortable: true },
  {
    id: "granted", header: "Status", accessorKey: "granted",
    cell: (row) => <NexusBadge variant={row.granted ? "success" : "danger"}>{row.granted ? "Granted" : "Denied"}</NexusBadge>,
  },
];

const schema = z.object({
  role_id: z.string().min(1, "Role is required"),
  access_right_id: z.string().min(1, "Access right is required"),
  granted: z.boolean(),
});

const fields: FieldDef[] = [
  {
    name: "role_id", label: "Role", type: "select", required: true,
    options: [{ label: "Admin", value: "1" }, { label: "Editor", value: "2" }, { label: "Viewer", value: "3" }],
  },
  {
    name: "access_right_id", label: "Access Right", type: "select", required: true,
    options: [
      { label: "Read Users", value: "1" }, { label: "Write Users", value: "2" },
      { label: "Delete Users", value: "3" }, { label: "Read Projects", value: "4" },
      { label: "Manage Roles", value: "5" },
    ],
  },
  { name: "granted", label: "Granted", type: "switch" },
];

export default function PermissionsPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [editItem, setEditItem] = useState<Permission | null>(null);
  const [deleteItem, setDeleteItem] = useState<Permission | null>(null);

  const { data: response, isLoading, isError } = usePermissions();
  const createPermission = useCreatePermission();
  const updatePermission = useUpdatePermission();
  const deletePermission = useDeletePermission();

  const permissions: Permission[] = useMemo(() => {
    if (response?.data) return response.data as Permission[];
    if (isError) return mockData;
    return mockData;
  }, [response, isError]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Permissions"
        description="Map access rights to roles."
        actions={<NexusButton onClick={() => setCreateOpen(true)}><Plus className="h-4 w-4" /> Assign Permission</NexusButton>}
      />

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-14 w-full rounded-lg" />)}
        </div>
      ) : (
        <CrudTable columns={columns} data={permissions} onEdit={setEditItem} onDelete={setDeleteItem} />
      )}

      <CrudFormDialog
        open={createOpen} onOpenChange={setCreateOpen}
        title="Assign Permission" description="Grant or deny an access right to a role."
        fields={fields} schema={schema}
        onSubmit={async (v) => { await createPermission.mutateAsync(v as any); setCreateOpen(false); }}
        submitLabel="Assign"
      />
      <CrudFormDialog
        open={!!editItem} onOpenChange={(o) => !o && setEditItem(null)}
        title="Edit Permission" description="Update permission assignment."
        fields={fields} schema={schema} initialValues={editItem || undefined}
        onSubmit={async (v) => { if (editItem) { await updatePermission.mutateAsync({ id: editItem.id, data: v as any }); setEditItem(null); } }}
        submitLabel="Save Changes"
      />
      <DeleteDialog
        open={!!deleteItem} onOpenChange={(o) => !o && setDeleteItem(null)}
        resourceName="Permission" itemName={`${deleteItem?.role_name} → ${deleteItem?.access_right_name}`}
        onConfirm={async () => { if (deleteItem) { await deletePermission.mutateAsync(String(deleteItem.id)); setDeleteItem(null); } }}
      />
    </div>
  );
}
