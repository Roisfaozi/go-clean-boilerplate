import { PageHeader } from "@/components/layout/page-header";
import { NexusButton } from "@casbin/ui";
import { NexusBadge } from "@casbin/ui";
import {
  NexusCard,
  NexusCardHeader,
  NexusCardTitle,
  NexusCardContent,
  NexusCardDescription,
} from "@casbin/ui";
import { DashboardGrid } from "@/components/layout/dashboard-grid";
import { Plus, FolderKanban, MoreHorizontal } from "lucide-react";

const mockProjects = [
  {
    id: "1",
    name: "Marketing Dashboard",
    slug: "marketing",
    description: "Analytics for marketing team",
    status: "active",
  },
  {
    id: "2",
    name: "Customer Portal",
    slug: "customer-portal",
    description: "Self-service customer portal",
    status: "active",
  },
  {
    id: "3",
    name: "Internal Tools",
    slug: "internal",
    description: "Developer productivity tools",
    status: "draft",
  },
  {
    id: "4",
    name: "API Gateway",
    slug: "api-gw",
    description: "Central API management",
    status: "archived",
  },
];

const statusVariant = (s: string) => {
  if (s === "active") return "success" as const;
  if (s === "draft") return "warning" as const;
  return "neutral" as const;
};

export default function ProjectsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Projects"
        description="Manage your workspace projects."
        actions={
          <NexusButton>
            <Plus className="h-4 w-4" />
            New Project
          </NexusButton>
        }
      />
      <DashboardGrid columns={3}>
        {mockProjects.map((p) => (
          <NexusCard
            key={p.id}
            className="flex flex-col justify-between transition-shadow hover:shadow-md"
          >
            <NexusCardHeader>
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-2">
                  <FolderKanban className="text-primary h-5 w-5" />
                  <NexusCardTitle>{p.name}</NexusCardTitle>
                </div>
                <NexusButton variant="ghost" size="icon" className="h-7 w-7">
                  <MoreHorizontal className="h-4 w-4" />
                </NexusButton>
              </div>
              <NexusCardDescription>{p.description}</NexusCardDescription>
            </NexusCardHeader>
            <NexusCardContent>
              <NexusBadge variant={statusVariant(p.status)}>
                {p.status}
              </NexusBadge>
            </NexusCardContent>
          </NexusCard>
        ))}
      </DashboardGrid>
    </div>
  );
}
