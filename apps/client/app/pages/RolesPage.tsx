import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { DataTable } from "@/components/patterns/data-table";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import {
  NexusCard,
  NexusCardHeader,
  NexusCardTitle,
  NexusCardContent,
} from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { NexusTextarea } from "@casbin/ui";
import { Plus, MoreHorizontal, X } from "lucide-react";

const mockRoles = [
  {
    id: "1",
    name: "Admin",
    description: "Full access to all resources",
    created_at: 1700000000,
  },
  {
    id: "2",
    name: "Editor",
    description: "Can create and edit content",
    created_at: 1700100000,
  },
  {
    id: "3",
    name: "Viewer",
    description: "Read-only access",
    created_at: 1700200000,
  },
  {
    id: "4",
    name: "Moderator",
    description: "Can moderate user content",
    created_at: 1700300000,
  },
];

export default function RolesPage() {
  const [showCreate, setShowCreate] = useState(false);

  const columns = [
    { key: "name", header: "Role Name", sortable: true },
    { key: "description", header: "Description" },
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
        title="Roles"
        description="Define and manage access roles."
        actions={
          <NexusButton onClick={() => setShowCreate(!showCreate)}>
            {showCreate ? (
              <X className="h-4 w-4" />
            ) : (
              <Plus className="h-4 w-4" />
            )}
            {showCreate ? "Cancel" : "Create Role"}
          </NexusButton>
        }
      />

      {showCreate && (
        <NexusCard className="animate-fade-in">
          <NexusCardHeader>
            <NexusCardTitle>New Role</NexusCardTitle>
          </NexusCardHeader>
          <NexusCardContent className="space-y-4">
            <FormGroup label="Role Name" required>
              <NexusInput placeholder="e.g. Manager" />
            </FormGroup>
            <FormGroup label="Description">
              <NexusTextarea placeholder="Describe the role's permissions..." />
            </FormGroup>
            <NexusButton>Create Role</NexusButton>
          </NexusCardContent>
        </NexusCard>
      )}

      <DataTable columns={columns} data={mockRoles as any} />
    </div>
  );
}
