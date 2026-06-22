import { PageHeader } from "@/components/layout/page-header";
import {
  CrudFormDialog,
  CrudTable,
  DeleteDialog,
  type CrudColumnDef,
  type FieldDef,
} from "@/features/shared";
import type { AccessRight } from "@/lib/api/types";
import { NexusBadge, NexusButton } from "@casbin/ui";
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
    id: "resource",
    header: "Resource",
    accessorKey: "resource",
    sortable: true,
    minWidth: 180,
  },
  {
    id: "action",
    header: "Action",
    accessorKey: "action",
    sortable: true,
    filterable: true,
    filterOptions: [
      { label: "Create", value: "create" },
      { label: "Read", value: "read" },
      { label: "Update", value: "update" },
      { label: "Delete", value: "delete" },
    ],
    cell: (row) => <NexusBadge variant="neutral">{row.action}</NexusBadge>,
  },
];

const createSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  resource: z.string().trim().min(1, "Resource is required").max(100),
  action: z.string().trim().min(1, "Action is required").max(50),
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
    name: "resource",
    label: "Resource",
    type: "text",
    required: true,
    placeholder: "e.g. users",
  },
  {
    name: "action",
    label: "Action",
    type: "select",
    required: true,
    options: [
      { label: "Create", value: "create" },
      { label: "Read", value: "read" },
      { label: "Update", value: "update" },
      { label: "Delete", value: "delete" },
    ],
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
