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
  useOrganizations,
  useCreateOrganization,
  useUpdateOrganization,
  useDeleteOrganization,
} from "./organizationHooks";
import type { Organization } from "@/lib/api/schemas";
import { Skeleton } from "@casbin/ui";

interface OrgRow extends Organization {
  id: string;
  members: number;
}

const mockOrgs: OrgRow[] = [
  {
    id: "1",
    name: "Acme Corp",
    slug: "acme",
    status: "active",
    owner_id: "u1",
    members: 45,
    created_at: 1700000000,
  },
  {
    id: "2",
    name: "Globex Inc",
    slug: "globex",
    status: "active",
    owner_id: "u2",
    members: 23,
    created_at: 1700100000,
  },
  {
    id: "3",
    name: "Initech",
    slug: "initech",
    status: "suspended",
    owner_id: "u3",
    members: 12,
    created_at: 1700200000,
  },
  {
    id: "4",
    name: "Umbrella Corp",
    slug: "umbrella",
    status: "active",
    owner_id: "u4",
    members: 67,
    created_at: 1700300000,
  },
  {
    id: "5",
    name: "Stark Industries",
    slug: "stark",
    status: "active",
    owner_id: "u5",
    members: 150,
    created_at: 1700400000,
  },
];

const columns: CrudColumnDef<OrgRow>[] = [
  {
    id: "name",
    header: "Organization",
    accessorKey: "name",
    sortable: true,
    minWidth: 180,
  },
  { id: "slug", header: "Slug", accessorKey: "slug" },
  { id: "members", header: "Members", accessorKey: "members", sortable: true },
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
];

const createSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  slug: z
    .string()
    .trim()
    .min(1, "Slug is required")
    .max(50)
    .regex(/^[a-z0-9-]+$/, "Lowercase letters, numbers, hyphens only"),
});

const editSchema = z.object({
  name: z.string().trim().min(1, "Name is required").max(100),
  slug: z
    .string()
    .trim()
    .min(1, "Slug is required")
    .max(50)
    .regex(/^[a-z0-9-]+$/, "Lowercase letters, numbers, hyphens only"),
  status: z.string().min(1, "Status is required"),
});

const createFields: FieldDef[] = [
  {
    name: "name",
    label: "Name",
    type: "text",
    required: true,
    placeholder: "e.g. Acme Corp",
  },
  {
    name: "slug",
    label: "Slug",
    type: "text",
    required: true,
    placeholder: "e.g. acme-corp",
    description: "URL-friendly identifier",
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
      { label: "Suspended", value: "suspended" },
    ],
  },
];

export default function OrganizationsPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [editItem, setEditItem] = useState<OrgRow | null>(null);
  const [deleteItem, setDeleteItem] = useState<OrgRow | null>(null);

  const { data: orgsResponse, isLoading, isError } = useOrganizations();
  const createOrg = useCreateOrganization();
  const updateOrg = useUpdateOrganization();
  const deleteOrg = useDeleteOrganization();

  const orgs: OrgRow[] = useMemo(() => {
    if (orgsResponse?.data) return orgsResponse.data as OrgRow[];
    if (isError) return mockOrgs;
    return mockOrgs;
  }, [orgsResponse, isError]);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Organizations"
        description="Manage organizations and their members."
        actions={
          <NexusButton onClick={() => setCreateOpen(true)}>
            <Plus className="h-4 w-4" /> New Organization
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
          data={orgs}
          selectable
          onEdit={setEditItem}
          onDelete={setDeleteItem}
          bulkActions={[
            {
              label: "Delete Selected",
              onClick: (ids) => {
                ids.forEach((id) => {
                  deleteOrg.mutate(String(id));
                });
              },
            },
          ]}
        />
      )}

      <CrudFormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        title="Create Organization"
        description="Add a new organization."
        fields={createFields}
        schema={createSchema}
        onSubmit={(values) => {
          createOrg.mutate(values as any);
          setCreateOpen(false);
        }}
        submitLabel="Create"
      />

      <CrudFormDialog
        open={!!editItem}
        onOpenChange={(open) => !open && setEditItem(null)}
        title="Edit Organization"
        description="Update organization details."
        fields={editFields}
        schema={editSchema}
        initialValues={editItem || undefined}
        onSubmit={(values) => {
          if (editItem) {
            updateOrg.mutate({
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
        resourceName="Organization"
        itemName={deleteItem?.name}
        onConfirm={() => {
          if (deleteItem) {
            deleteOrg.mutate(String(deleteItem.id));
            setDeleteItem(null);
          }
        }}
      />
    </div>
  );
}
