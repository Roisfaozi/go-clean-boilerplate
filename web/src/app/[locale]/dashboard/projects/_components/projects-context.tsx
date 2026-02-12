"use client";

import { createContext, useContext, useCallback, ReactNode } from "react";
import { Project, projectsApi } from "~/lib/api/projects";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";
import useSWR from "swr";

interface ProjectsContextType {
  projects: Project[];
  isLoading: boolean;
  fetchProjects: () => Promise<void>;
  createProject: (data: { name: string; domain: string }) => Promise<void>;
  updateProject: (id: string, data: any) => Promise<void>;
  deleteProject: (id: string) => Promise<void>;
}

const ProjectsContext = createContext<ProjectsContextType | undefined>(
  undefined
);

export function ProjectsProvider({
  children,
  initialData,
}: {
  children: ReactNode;
  initialData?: Project[];
}) {
  const { currentOrganization } = useOrganizationStore();

  const {
    data: projects = [],
    isLoading,
    mutate,
  } = useSWR(
    currentOrganization ? ["/api/v1/projects", currentOrganization.id] : null,
    ([_, orgId]) => projectsApi.getAll(orgId),
    {
      fallbackData: initialData,
      keepPreviousData: true,
    }
  );

  const fetchProjects = useCallback(async () => {
    await mutate();
  }, [mutate]);

  const createProject = useCallback(
    async (data: { name: string; domain: string }) => {
      if (!currentOrganization) return;
      try {
        await projectsApi.create(currentOrganization.id, data);
        toast.success("Project created successfully");
        await mutate();
      } catch (error) {
        toast.error("Failed to create project");
        throw error;
      }
    },
    [currentOrganization, mutate]
  );

  const updateProject = useCallback(
    async (id: string, data: any) => {
      if (!currentOrganization) return;
      try {
        await projectsApi.update(currentOrganization.id, id, data);
        toast.success("Project updated successfully");
        await mutate();
      } catch (error) {
        toast.error("Failed to update project");
      }
    },
    [currentOrganization, mutate]
  );

  const deleteProject = useCallback(
    async (id: string) => {
      if (!currentOrganization) return;
      try {
        await projectsApi.delete(currentOrganization.id, id);
        toast.success("Project deleted successfully");
        await mutate();
      } catch (error) {
        toast.error("Failed to delete project");
      }
    },
    [currentOrganization, mutate]
  );

  return (
    <ProjectsContext.Provider
      value={{
        projects,
        isLoading,
        fetchProjects,
        createProject,
        updateProject,
        deleteProject,
      }}
    >
      {children}
    </ProjectsContext.Provider>
  );
}

export function useProjects() {
  const context = useContext(ProjectsContext);
  if (context === undefined) {
    throw new Error("useProjects must be used within a ProjectsProvider");
  }
  return context;
}
