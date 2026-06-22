import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { projectService } from "./projectService";
import { toast } from "@casbin/ui";
import type { Project } from "@/lib/api/schemas";

const KEY = ["projects"];

export function useProjects(params?: {
  page?: number;
  limit?: number;
  org_id?: string;
}) {
  return useQuery({
    queryKey: [...KEY, params],
    queryFn: () => projectService.list(params),
  });
}

export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: projectService.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Project created");
    },
    onError: () => toast.error("Failed to create project"),
  });
}

export function useUpdateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Project> }) =>
      projectService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Project updated");
    },
    onError: () => toast.error("Failed to update project"),
  });
}

export function useDeleteProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => projectService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Project deleted");
    },
    onError: () => toast.error("Failed to delete project"),
  });
}
