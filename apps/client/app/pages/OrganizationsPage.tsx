import { PageHeader } from "@/components/layout/page-header";
import { DataTable } from "@/components/patterns/data-table";
import { NexusBadge } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { Plus, MoreHorizontal } from "lucide-react";

const mockOrgs = [
  {
    id: "1",
    name: "Acme Corp",
    slug: "acme",
    status: "active",
    owner_id: "u1",
    members: 45,
  },
  {
    id: "2",
    name: "Globex Inc",
    slug: "globex",
    status: "active",
    owner_id: "u2",
    members: 23,
  },
  {
    id: "3",
    name: "Initech",
    slug: "initech",
    status: "suspended",
    owner_id: "u3",
    members: 12,
  },
];

export default function OrganizationsPage() {
  const columns = [
    { key: "name", header: "Organization", sortable: true },
    { key: "slug", header: "Slug" },
    { key: "members", header: "Members", sortable: true },
    {
      key: "status",
      header: "Status",
      render: (o: (typeof mockOrgs)[0]) => (
        <NexusBadge variant={o.status === "active" ? "success" : "danger"}>
          {o.status}
        </NexusBadge>
      ),
    },
    {
      key: "actions",
      header: "",
      className: "w-12",
      render: () => (
        <NexusButton variant="ghost" size="icon" className="h-8 w-8">
          <MoreHorizontal className="h-4 w-4" />
        </NexusButton>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <PageHeader
        title="Organizations"
        description="Manage organizations and their members."
        actions={
          <NexusButton>
            <Plus className="h-4 w-4" />
            New Organization
          </NexusButton>
        }
      />
      <DataTable columns={columns} data={mockOrgs as any} />
    </div>
  );
}
