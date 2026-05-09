import { PageHeader } from "@/components/layout/page-header";
import { SidebarGroup, SidebarItem, SidebarCollapsible } from "@/components/navigation/sidebar-nav";
import { MegaDropdown } from "@/components/navigation/mega-dropdown";
import {  NexusButton  } from "@casbin/ui";
import { Activity, Users, Settings, FileText, Mail, GitBranch } from "lucide-react";

export default function ShowcaseNavigation() {
  return (
    <div className="space-y-10 max-w-5xl">
      <PageHeader title="Navigation" description="Sidebar navigation, collapsible groups, and mega dropdown." />

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <section className="space-y-4">
          <h2 className="text-h2 text-foreground">Sidebar Navigation</h2>
          <div className="border border-border rounded-lg p-3 max-w-[240px]">
            <SidebarGroup label="Main">
              <SidebarItem label="Dashboard" href="/" icon={<Activity className="h-5 w-5" />} />
              <SidebarItem label="Users" href="/users" icon={<Users className="h-5 w-5" />} />
            </SidebarGroup>
            <div className="mt-2">
              <SidebarCollapsible label="Settings" icon={<Settings className="h-5 w-5" />} defaultOpen>
                <SidebarItem label="General" href="/settings" />
                <SidebarItem label="Security" href="/settings" />
              </SidebarCollapsible>
            </div>
          </div>
        </section>

        <section className="space-y-4">
          <h2 className="text-h2 text-foreground">Mega Dropdown</h2>
          <MegaDropdown
            trigger={<NexusButton variant="outline">Products ▾</NexusButton>}
            sections={[
              {
                title: "Platform",
                items: [
                  { label: "Analytics", description: "Track metrics", icon: <Activity className="h-4 w-4" /> },
                  { label: "Automation", description: "Workflows", icon: <GitBranch className="h-4 w-4" /> },
                ],
              },
              {
                title: "Resources",
                items: [
                  { label: "Documentation", description: "Guides & API", icon: <FileText className="h-4 w-4" /> },
                  { label: "Support", description: "Get help", icon: <Mail className="h-4 w-4" /> },
                ],
              },
            ]}
          />
        </section>
      </div>
    </div>
  );
}
