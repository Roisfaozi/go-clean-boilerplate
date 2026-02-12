"use client";

import { createContext, useContext, useState, useCallback, ReactNode, useEffect } from "react";
import { Project, projectsApi } from "~/lib/api/projects";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";

interface ProjectsContextType {
  projects: Project[];
  isLoading: boolean;
  fetchProjects: () => Promise<void>;
  createProject: (data: { name: string; domain: string }) => Promise<void>;
  updateProject: (id: string, data: any) => Promise<void>;
  deleteProject: (id: string) => Promise<void>;
}

const ProjectsContext = createContext<ProjectsContextType | undefined>(undefined);

export function ProjectsProvider({ children }: { children: ReactNode }) {
  const { currentOrganization } = useOrganizationStore();
  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const fetchProjects = useCallback(async () => {
    if (!currentOrganization) return;
    setIsLoading(true);
    try {
      const data = await projectsApi.getAll(currentOrganization.id);
      setProjects(data || []);
    } catch (error) {
      console.error("Failed to fetch projects", error);
      toast.error("Failed to load projects");
    } finally {
      setIsLoading(false);
    }
  }, [currentOrganization]);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  const createProject = useCallback(async (data: { name: string; domain: string }) => {
    if (!currentOrganization) return;
    try {
      await projectsApi.create(currentOrganization.id, data);
      toast.success("Project created successfully");
      await fetchProjects();
    } catch (error) {
      toast.error("Failed to create project");
      throw error;
    }
  }, [currentOrganization, fetchProjects]);

  const updateProject = useCallback(async (id: string, data: any) => {
    if (!currentOrganization) return;
    try {
      await projectsApi.update(currentOrganization.id, id, data);
      toast.success("Project updated successfully");
      await fetchProjects();
    } catch (error) {
      toast.error("Failed to update project");
    }
  }, [currentOrganization, fetchProjects]);

  const deleteProject = useCallback(async (id: string) => {
    if (!currentOrganization) return;
    try {
      await projectsApi.delete(currentOrganization.id, id);
      toast.success("Project deleted successfully");
      await fetchProjects();
    } catch (error) {
      toast.error("Failed to delete project");
    }
  }, [currentOrganization, fetchProjects]);

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
