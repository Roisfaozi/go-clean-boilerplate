import { DashboardHeader } from "~/components/layout/dashboard/header";
import { Sidebar } from "~/components/layout/sidebar";
import { DashboardShellProvider } from "./_components/dashboard-shell-context";
import { organizationsApi } from "~/lib/api/organizations";

// Better way: import usePresence in a dedicated client component or just here if we make this function client
// Let's create a Client wrapper for the layout logic
import { DashboardLayoutClient } from "./layout-client";

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  // 1. Fetch organizations on Server (Critical for Navigation/Switcher)
  let initialOrgs = undefined;
  try {
    const resp = await organizationsApi.getMyOrganizations();
    initialOrgs = resp.data?.organizations;
  } catch (error) {
    console.error("Failed to fetch initial orgs on server", error);
  }

  return (
    <DashboardShellProvider initialData={initialOrgs}>
      <DashboardLayoutClient>{children}</DashboardLayoutClient>
    </DashboardShellProvider>
  );
}
