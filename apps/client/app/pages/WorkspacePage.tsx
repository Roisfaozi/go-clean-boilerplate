import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@casbin/ui";
import { OrganizationSwitcher } from "@/features/organizations/organization-switcher";
import { MembersTableUI } from "@/features/organizations/members-table";
import { InviteMemberModal } from "@/features/organizations/invite-member-modal";
import { NexusCard } from "@casbin/ui";
import { Badge } from "@casbin/ui";
import { Users, FolderKanban, Shield } from "lucide-react";

const stats = [
  { label: "Members", value: "6", icon: Users, change: "+2 this month" },
  { label: "Projects", value: "4", icon: FolderKanban, change: "+1 this week" },
  { label: "Roles", value: "4", icon: Shield, change: "No change" },
];

export default function WorkspacePage() {
  const [tab, setTab] = useState("members");

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <PageHeader
          title="Workspace"
          description="Manage your organization, members, and settings"
        />
        <div className="flex items-center gap-3">
          <div className="w-56">
            <OrganizationSwitcher />
          </div>
          <InviteMemberModal />
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {stats.map((stat) => (
          <NexusCard key={stat.label} className="p-4">
            <div className="flex items-center gap-3">
              <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded-md">
                <stat.icon className="text-primary h-5 w-5" />
              </div>
              <div>
                <p className="text-foreground text-2xl font-bold">
                  {stat.value}
                </p>
                <p className="text-muted-foreground text-xs">{stat.label}</p>
              </div>
              <Badge variant="outline" className="ml-auto text-[10px]">
                {stat.change}
              </Badge>
            </div>
          </NexusCard>
        ))}
      </div>

      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          <TabsTrigger value="members">Members</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>

        <TabsContent value="members" className="mt-4">
          <MembersTableUI />
        </TabsContent>

        <TabsContent value="settings" className="mt-4">
          <NexusCard className="p-6">
            <h3 className="text-foreground mb-4 text-sm font-semibold">
              Organization Settings
            </h3>
            <div className="space-y-3 text-sm">
              <div className="border-border flex justify-between border-b py-2">
                <span className="text-muted-foreground">Name</span>
                <span className="text-foreground font-medium">Acme Corp</span>
              </div>
              <div className="border-border flex justify-between border-b py-2">
                <span className="text-muted-foreground">Slug</span>
                <span className="text-foreground font-mono text-xs">acme</span>
              </div>
              <div className="border-border flex justify-between border-b py-2">
                <span className="text-muted-foreground">Plan</span>
                <Badge>Enterprise</Badge>
              </div>
              <div className="flex justify-between py-2">
                <span className="text-muted-foreground">Created</span>
                <span className="text-foreground">Jan 15, 2024</span>
              </div>
            </div>
          </NexusCard>
        </TabsContent>
      </Tabs>
    </div>
  );
}
