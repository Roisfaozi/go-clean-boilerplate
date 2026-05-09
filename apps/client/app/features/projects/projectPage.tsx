import { useState } from "react";
import { z } from "zod";
import { PageHeader } from "@/components/layout/page-header";
import {  NexusButton  } from "@casbin/ui";
import {  NexusBadge  } from "@casbin/ui";
import { CrudTable, CrudFormDialog, DeleteDialog, type CrudColumnDef, type FieldDef } from "@/features/shared";
import { Plus } from "lucide-react";

interface ProjectRow {
  id: string;
  name: string;
  slug: string;
  description: string;
  status: string;
  organization_id: string;
}

const mockProjects: ProjectRow[] = [
  { id: "1", name: "Marketing Dashboard", slug: "marketing", description: "Analytics for marketing team", status: "active", organization_id: "1" },
  { id: "2", name: "Customer Portal", slug: "customer-portal", description: "Self-service customer portal", status: "active", organization_id: "1" },
  { id: "3", name: "Internal Tools", slug: "internal", description: "Developer productivity tools", status: "draft", organization_id: "2" },
  { id: "4", name: "API Gateway", slug: "api-gw", description: "Central API management", status: "archived", organization_id: "2" },
  { id: "5", name: "Mobile App", slug: "mobile-app", description: "Cross-platform mobile application", status: "active", organization_id: "1" },
];

const statusVariant = (s: string) => s === "active" ? "success" as const : s === "draft" ? "warning" as const : "neutral" as const;

const columns: CrudColumnDef<ProjectRow>[] = [
  { id: "name", header: "Project", accessorKey: "name", sortable: true, minWidth: 180 },
  { id: "slug", header: "Slug", accessorKey: "slug" },
  { id: "description", header: "Description", accessorKey: "description", minWidth: 200 },
  {
    id: "status", header: "Status", accessorKey: "status", sortable: true,
    filterable: true,
    filterOptions: [{ label: "Active", value: "active" }, { label: "Draft", value: "draft" }, { label: "Archived", value: "archived" }],
    cell: (row) => <NexusBadge variant={statusVariant(row.status)}>{row.status}</NexusBadge>,
  },
];

const createSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  slug: z.string().trim().min(1, "Slug is required").max(50).regex(/^[a-z0-9-]+$/, "Lowercase, numbers, hyphens"),
  description: z.string().max(500).optional(),
  organization_id: z.string().min(1, "Organization is required"),
});

const editSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  slug: z.string().trim().min(1).max(50).regex(/^[a-z0-9-]+$/),
  description: z.string().max(500).optional(),
  status: z.string().min(1, "Status is required"),
});

const createFields: FieldDef[] = [
  { name: "name", label: "Name", type: "text", required: true, placeholder: "e.g. Marketing Dashboard" },
  { name: "slug", label: "Slug", type: "text", required: true, placeholder: "e.g. marketing-dashboard" },
  { name: "description", label: "Description", type: "textarea", placeholder: "Project description…" },
  {
    name: "organization_id", label: "Organization", type: "select", required: true,
    options: [{ label: "Acme Corp", value: "1" }, { label: "Globex Inc", value: "2" }],
  },
];

const editFields: FieldDef[] = [
  { name: "name", label: "Name", type: "text", required: true },
  { name: "slug", label: "Slug", type: "text", required: true },
  { name: "description", label: "Description", type: "textarea" },
  {
    name: "status", label: "Status", type: "select", required: true,
    options: [{ label: "Active", value: "active" }, { label: "Draft", value: "draft" }, { label: "Archived", value: "archived" }],
  },
];

export default function ProjectsPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [editItem, setEditItem] = useState<ProjectRow | null>(null);
  const [deleteItem, setDeleteItem] = useState<ProjectRow | null>(null);

  return (
    <div className="space-y-6">
      <PageHeader title="Projects" description="Manage your workspace projects." actions={<NexusButton onClick={() => setCreateOpen(true)}><Plus className="h-4 w-4" /> New Project</NexusButton>} />
      <CrudTable columns={columns} data={mockProjects} onEdit={setEditItem} onDelete={setDeleteItem} selectable />
      <CrudFormDialog open={createOpen} onOpenChange={setCreateOpen} title="Create Project" fields={createFields} schema={createSchema} onSubmit={async (v) => console.log("create", v)} submitLabel="Create Project" />
      <CrudFormDialog open={!!editItem} onOpenChange={(o) => !o && setEditItem(null)} title="Edit Project" fields={editFields} schema={editSchema} initialValues={editItem || undefined} onSubmit={async (v) => console.log("update", editItem?.id, v)} submitLabel="Save Changes" />
      <DeleteDialog open={!!deleteItem} onOpenChange={(o) => !o && setDeleteItem(null)} resourceName="Project" itemName={deleteItem?.name} onConfirm={async () => console.log("delete", deleteItem?.id)} />
    </div>
  );
}
