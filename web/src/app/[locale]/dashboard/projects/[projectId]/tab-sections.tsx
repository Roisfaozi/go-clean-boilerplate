"use client";

import { Project } from "~/lib/api/projects";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import { ProjectDetailProvider } from "./_components/project-detail-context";
import { ProjectDetailsForm } from "./_components/project-details-form";
import { ProjectDangerZone } from "./_components/project-danger-zone";

export default function TabSections({ project }: { project: Project }) {
  return (
    <ProjectDetailProvider initialProject={project}>
      <Tabs defaultValue="details" className="space-y-6">
        <TabsList>
          <TabsTrigger value="details">Details</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>
        
        <TabsContent value="details">
          <ProjectDetailsForm />
        </TabsContent>
        
        <TabsContent value="settings">
          <ProjectDangerZone />
        </TabsContent>
      </Tabs>
    </ProjectDetailProvider>
  );
}
