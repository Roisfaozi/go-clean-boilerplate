import { PageHeader } from "@/components/layout/page-header";
import {
  CrudFormDialog,
  CrudTable,
  DeleteDialog,
  type CrudColumnDef,
  type FieldDef,
} from "@/features/shared";
import type { Resource } from "@/lib/api/types";
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

const mockData: Resource[] = [
  {
    id: "1",
    name: "Users",
    slug: "users",
    description: "User management resource",
    status: "active",
    created_at: 1700000000,
  },
  {
    id: "2",
    name: "Projects",
    slug: "projects",
    description: "Project management resource",
    status: "active",
    created_at: 1700100000,
  },
  {
    id: "3",
    name: "Roles",
    slug: "roles",
    description: "Role management resource",
    status: "active",
    created_at: 1700200000,
  },
  {
    id: "4",
    name: "Organizations",
    slug: "organizations",
    description: "Organization management",
    status: "active",
    created_at: 1700300000,
  },
  {
    id: "5",
    name: "Audit Logs",
    slug: "audit-logs",
    description: "System audit trail",
    status: "active",
    created_at: 1700400000,
  },
  {
    id: "6",
    name: "Settings",
    slug: "settings",
    description: "Application settings",
    status: "inactive",
    created_at: 1700500000,
  },
];

const columns: CrudColumnDef<Resource>[] = [
  { id: "name", header: "Name", accessorKey: "name", sortable: true, minWidth: 160 },
  { id: "description", header: "Description", accessorKey: "description", minWidth: 300 },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    filterable: true,
    filterOptions: [
      { label: "Active", value: "active" },
      { label: "Inactive", value: "inactive" },
    ],
    cell: (row) => (
      <NexusBadge variant={row.status === "inactive" ? "neutral" : "success"}>
        {row.status ?? "active"}
      </NexusBadge>
    ),
  },
];

const createSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  description: z.string().max(500).optional(),
});

const editSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  description: z.string().max(500).optional(),
  status: z.string().min(1, "Status is required"),
});

const createFields: FieldDef[] = [
  { name: "name", label: "Resource Name", type: "text", required: true, placeholder: "e.g. Users" },
  {
    name: "description",
    label: "Description",
    type: "textarea",
    placeholder: "Describe this resource…",
  },
];

const editFields: FieldDef[] = [
  ...createFields,
  {
    name: "status",
    label: "Status",
    type: "select",
    required: true,
    options: [
      { label: "Active", value: "active" },
      { label: "Inactive", value: "inactive" },
    ],
  },
];

export default function ResourcesPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [editItem, setEditItem] = useState<Resource | null>(null);
  const [deleteItem, setDeleteItem] = useState<Resource | null>(null);

  const { data: response, isLoading, refetch } = useResources();
  const createResource = useCreateResource();
  const updateResource = useUpdateResource();
  const deleteResource = useDeleteResource();

  const resources: Resource[] = useMemo(() => {
    if (response?.data) return response.data as Resource[];
    return [];
  }, [response]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Resources"
        description="Register and manage API resources (Access Rights)."
        actions={
          <div className="flex gap-2">
            <NexusButton variant="outline" onClick={() => refetch()}>
              Refresh
            </NexusButton>
            <NexusButton onClick={() => setCreateOpen(true)}>
              <Plus className="mr-2 h-4 w-4" /> New Resource
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
        title="Register Resource"
        description="Add a new API resource."
        fields={createFields}
        schema={createSchema}
        loading={createResource.isPending}
        onSubmit={async (v) => {
          await createResource.mutateAsync(v as any);
          setCreateOpen(false);
        }}
        submitLabel="Create Resource"
      />
      <CrudFormDialog
        open={!!editItem}
        onOpenChange={(o) => !o && setEditItem(null)}
        title="Edit Resource"
        description="Update resource details."
        fields={editFields}
        schema={editSchema}
        loading={updateResource.isPending}
        initialValues={editItem || undefined}
        onSubmit={async (v) => {
          if (editItem) {
            await updateResource.mutateAsync({ id: editItem.id, data: v as any });
            setEditItem(null);
          }
        }}
        submitLabel="Save Changes"
      />
      <DeleteDialog
        open={!!deleteItem}
        onOpenChange={(o) => !o && setDeleteItem(null)}
        resourceName="Resource"
        itemName={deleteItem?.name}
        loading={deleteResource.isPending}
        onConfirm={async () => {
          if (deleteItem) {
            await deleteResource.mutateAsync(String(deleteItem.id));
            setDeleteItem(null);
          }
        }}
      />
    </div>
  );
}
