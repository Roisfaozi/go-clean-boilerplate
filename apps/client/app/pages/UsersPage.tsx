import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { DataTable } from "@/components/patterns/data-table";
import { NexusBadge } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { Plus, Search, MoreHorizontal } from "lucide-react";

const mockUsers = [
  {
    id: "1",
    name: "Alice Johnson",
    email: "alice@nexus.io",
    username: "alice",
    role: "Admin",
    status: "active",
    created_at: 1700000000,
  },
  {
    id: "2",
    name: "Bob Smith",
    email: "bob@nexus.io",
    username: "bob",
    role: "Editor",
    status: "active",
    created_at: 1700100000,
  },
  {
    id: "3",
    name: "Carol White",
    email: "carol@nexus.io",
    username: "carol",
    role: "Viewer",
    status: "suspended",
    created_at: 1700200000,
  },
  {
    id: "4",
    name: "David Brown",
    email: "david@nexus.io",
    username: "david",
    role: "Admin",
    status: "active",
    created_at: 1700300000,
  },
  {
    id: "5",
    name: "Eve Davis",
    email: "eve@nexus.io",
    username: "eve",
    role: "Editor",
    status: "active",
    created_at: 1700400000,
  },
];

export default function UsersPage() {
  const [search, setSearch] = useState("");
  const filtered = mockUsers.filter(
    (u) =>
      u.name.toLowerCase().includes(search.toLowerCase()) ||
      u.email.toLowerCase().includes(search.toLowerCase()),
  );

  const columns = [
    {
      key: "name",
      header: "Name",
      sortable: true,
      render: (u: (typeof mockUsers)[0]) => (
        <div>
          <p className="text-foreground font-medium">{u.name}</p>
          <p className="text-caption text-muted-foreground">{u.email}</p>
        </div>
      ),
    },
    { key: "username", header: "Username", sortable: true },
    { key: "role", header: "Role", sortable: true },
    {
      key: "status",
      header: "Status",
      render: (u: (typeof mockUsers)[0]) => (
        <NexusBadge variant={u.status === "active" ? "success" : "danger"}>
          {u.status}
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
        title="Users"
        description="Manage user accounts and permissions."
        actions={
          <NexusButton>
            <Plus className="h-4 w-4" />
            Add User
          </NexusButton>
        }
      />

      <div className="flex items-center gap-3">
        <div className="relative max-w-sm flex-1">
          <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <NexusInput
            placeholder="Search users..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="h-9 pl-10"
          />
        </div>
      </div>

      <DataTable columns={columns} data={filtered as any} />
    </div>
  );
}
