import { ProjectsProvider } from "./_components/projects-context";
import { ProjectsGrid } from "./_components/projects-grid";
import { projectsApi } from "~/lib/api/projects";
import { cookies } from "next/headers";

export default async function ProjectsPage() {
  const cookieStore = await cookies();
  const orgId = cookieStore.get("organization_id")?.value;

  // 1. Fetch initial data on Server if orgId exists
  let initialData = undefined;
  if (orgId) {
    try {
      initialData = await projectsApi.getAll(orgId);
    } catch (error) {
      console.error("Failed to fetch initial projects on server", error);
    }
  }

  return (
    <ProjectsProvider initialData={initialData}>
      <div className="space-y-6">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Projects</h2>
          <p className="text-muted-foreground">
            Manage your application environments and deployments.
          </p>
        </div>

        <ProjectsGrid />
      </div>
    </ProjectsProvider>
  );
}
