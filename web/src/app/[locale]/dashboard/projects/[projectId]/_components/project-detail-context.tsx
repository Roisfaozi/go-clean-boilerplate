"use client";

import {
  createContext,
  useContext,
  useState,
  useCallback,
  ReactNode,
} from "react";
import { Project, projectsApi } from "~/lib/api/projects";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

interface ProjectDetailContextType {
  project: Project;
  isLoading: boolean;
  updateProject: (data: any) => Promise<void>;
  deleteProject: () => Promise<void>;
}

const ProjectDetailContext = createContext<
  ProjectDetailContextType | undefined
>(undefined);

export function ProjectDetailProvider({
  initialProject,
  children,
}: {
  initialProject: Project;
  children: ReactNode;
}) {
  const { currentOrganization } = useOrganizationStore();
  const [project, setProject] = useState<Project>(initialProject);
  const [isLoading, setIsLoading] = useState(false);
  const router = useRouter();

  const updateProject = useCallback(
    async (data: any) => {
      if (!currentOrganization) return;
      setIsLoading(true);
      try {
        const resp = await projectsApi.update(project.id, data);
        if (resp) {
          setProject(resp);
          toast.success("Project updated successfully");
        }
      } catch (error) {
        toast.error("Failed to update project");
      } finally {
        setIsLoading(false);
      }
    },
    [currentOrganization, project.id]
  );

  const deleteProject = useCallback(async () => {
    if (!currentOrganization) return;
    if (
      !confirm(
        "Are you sure you want to delete this project? This action cannot be undone."
      )
    )
      return;

    setIsLoading(true);
    try {
      await projectsApi.delete(project.id);
      toast.success("Project deleted successfully");
      router.push("/dashboard/projects");
    } catch (error) {
      toast.error("Failed to delete project");
    } finally {
      setIsLoading(false);
    }
  }, [currentOrganization, project.id, router]);

  return (
    <ProjectDetailContext.Provider
      value={{
        project,
        isLoading,
        updateProject,
        deleteProject,
      }}
    >
      {children}
    </ProjectDetailContext.Provider>
  );
}

export function useProjectDetail() {
  const context = useContext(ProjectDetailContext);
  if (context === undefined) {
    throw new Error(
      "useProjectDetail must be used within a ProjectDetailProvider"
    );
  }
  return context;
}
