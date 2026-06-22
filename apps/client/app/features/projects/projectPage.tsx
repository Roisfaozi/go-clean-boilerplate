import { useState } from "react";
import { z } from "zod";
import { PageHeader } from "@/components/layout/page-header";
import { NexusButton, NexusBadge } from "@casbin/ui";
import {
  CrudTable,
  CrudFormDialog,
  DeleteDialog,
  type CrudColumnDef,
  type FieldDef,
} from "@/features/shared";
import { Plus, RefreshCcw } from "lucide-react";
import {
  useProjects,
  useCreateProject,
  useUpdateProject,
  useDeleteProject,
} from "./projectHooks";
import { useOrganizationStore } from "@/stores/organization-store";
import type { Project } from "@/lib/api/schemas";

const statusVariant = (s: string) =>
  s === "active"
    ? ("success" as const)
    : s === "draft"
      ? ("warning" as const)
      : ("neutral" as const);

const columns: CrudColumnDef<Project>[] = [
  {
    id: "name",
    header: "Project",
    accessorKey: "name",
    sortable: true,
    minWidth: 180,
  },
  { id: "slug", header: "Slug", accessorKey: "slug" },
  {
    id: "description",
    header: "Description",
    accessorKey: "description",
    minWidth: 200,
  },
  {
    id: "status",
    header: "Status",
    accessorKey: "status",
    sortable: true,
    filterable: true,
    filterOptions: [
      { label: "Active", value: "active" },
      { label: "Draft", value: "draft" },
      { label: "Archived", value: "archived" },
    ],
    cell: (row) => (
      <NexusBadge variant={statusVariant(row.status)}>{row.status}</NexusBadge>
    ),
  },
];

const createSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  slug: z
    .string()
    .trim()
    .min(1, "Slug is required")
    .max(50)
    .regex(/^[a-z0-9-]+$/, "Lowercase, numbers, hyphens"),
  description: z.string().max(500).optional(),
  organization_id: z.string().min(1, "Organization is required"),
});

const editSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  slug: z
    .string()
    .trim()
    .min(1)
    .max(50)
    .regex(/^[a-z0-9-]+$/),
  description: z.string().max(500).optional(),
  status: z.string().min(1, "Status is required"),
});

const createFields: FieldDef[] = [
  {
    name: "name",
    label: "Name",
    type: "text",
    required: true,
    placeholder: "e.g. Marketing Dashboard",
  },
  {
    name: "slug",
    label: "Slug",
    type: "text",
    required: true,
    placeholder: "e.g. marketing-dashboard",
  },
  {
    name: "description",
    label: "Description",
    type: "textarea",
    placeholder: "Project description…",
  },
];

const editFields: FieldDef[] = [
  { name: "name", label: "Name", type: "text", required: true },
  { name: "slug", label: "Slug", type: "text", required: true },
  { name: "description", label: "Description", type: "textarea" },
  {
    name: "status",
    label: "Status",
    type: "select",
    required: true,
    options: [
      { label: "Active", value: "active" },
      { label: "Draft", value: "draft" },
      { label: "Archived", value: "archived" },
    ],
  },
];

export default function ProjectsPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [editItem, setEditItem] = useState<Project | null>(null);
  const [deleteItem, setDeleteItem] = useState<Project | null>(null);

  const activeOrg = useOrganizationStore((s) => s.activeOrganization);
  const { data, isLoading, refetch } = useProjects({ org_id: activeOrg?.id });

  const createMutation = useCreateProject();
  const updateMutation = useUpdateProject();
  const deleteMutation = useDeleteProject();

  const handleCreate = (values: any) => {
    if (!activeOrg) return;
    createMutation.mutate({
      ...values,
      organization_id: activeOrg.id,
    });
    setCreateOpen(false);
  };

  const handleUpdate = (values: any) => {
    if (!editItem) return;
    updateMutation.mutate({
      id: editItem.id,
      data: values,
    });
    setEditItem(null);
  };

  const handleDelete = () => {
    if (!deleteItem) return;
    deleteMutation.mutate(deleteItem.id);
    setDeleteItem(null);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Projects"
        description="Manage your workspace projects."
        actions={
          <div className="flex gap-2">
            <NexusButton
              variant="outline"
              size="icon"
              onClick={() => refetch()}
            >
              <RefreshCcw className="h-4 w-4" />
            </NexusButton>
            <NexusButton onClick={() => setCreateOpen(true)}>
              <Plus className="h-4 w-4" /> New Project
            </NexusButton>
          </div>
        }
      />
      <CrudTable
        columns={columns}
        data={data?.data || []}
        loading={isLoading}
        onEdit={setEditItem}
        onDelete={setDeleteItem}
        selectable
      />
      <CrudFormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        title="Create Project"
        fields={createFields}
        schema={createSchema.omit({ organization_id: true })}
        onSubmit={handleCreate}
        submitLabel="Create Project"
        loading={createMutation.isPending}
      />
      <CrudFormDialog
        open={!!editItem}
        onOpenChange={(o) => !o && setEditItem(null)}
        title="Edit Project"
        fields={editFields}
        schema={editSchema}
        initialValues={editItem || undefined}
        onSubmit={handleUpdate}
        submitLabel="Save Changes"
        loading={updateMutation.isPending}
      />
      <DeleteDialog
        open={!!deleteItem}
        onOpenChange={(o) => !o && setDeleteItem(null)}
        resourceName="Project"
        itemName={deleteItem?.name}
        onConfirm={handleDelete}
        loading={deleteMutation.isPending}
      />
    </div>
  );
}
