import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import {  Tabs, TabsContent, TabsList, TabsTrigger  } from "@casbin/ui";
import { OrganizationSwitcher } from "@/features/organizations/organization-switcher";
import { MembersTableUI } from "@/features/organizations/members-table";
import { InviteMemberModal } from "@/features/organizations/invite-member-modal";
import {  NexusCard  } from "@casbin/ui";
import {  Badge  } from "@casbin/ui";
import { Users, Building2, FolderKanban, Shield } from "lucide-react";

const stats = [
  { label: "Members", value: "6", icon: Users, change: "+2 this month" },
  { label: "Projects", value: "4", icon: FolderKanban, change: "+1 this week" },
  { label: "Roles", value: "4", icon: Shield, change: "No change" },
];

export default function WorkspacePage() {
  const [tab, setTab] = useState("members");

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
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
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {stats.map((stat) => (
          <NexusCard key={stat.label} className="p-4">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-md bg-primary/10 flex items-center justify-center">
                <stat.icon className="h-5 w-5 text-primary" />
              </div>
              <div>
                <p className="text-2xl font-bold text-foreground">{stat.value}</p>
                <p className="text-xs text-muted-foreground">{stat.label}</p>
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
            <h3 className="text-sm font-semibold text-foreground mb-4">Organization Settings</h3>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between py-2 border-b border-border">
                <span className="text-muted-foreground">Name</span>
                <span className="font-medium text-foreground">Acme Corp</span>
              </div>
              <div className="flex justify-between py-2 border-b border-border">
                <span className="text-muted-foreground">Slug</span>
                <span className="font-mono text-xs text-foreground">acme</span>
              </div>
              <div className="flex justify-between py-2 border-b border-border">
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
