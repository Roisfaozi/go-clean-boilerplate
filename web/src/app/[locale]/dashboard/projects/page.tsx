"use client";

import { ProjectsProvider } from "./_components/projects-context";
import { ProjectsGrid } from "./_components/projects-grid";

export default function ProjectsPage() {
  return (
    <ProjectsProvider>
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
