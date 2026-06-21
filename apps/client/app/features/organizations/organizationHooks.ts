import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { organizationService } from "./organizationService";
import { toast } from "@casbin/ui";

const KEY = ["organizations"];

export function useOrganizations(params?: { page?: number; limit?: number }) {
  return useQuery({
    queryKey: [...KEY, params],
    queryFn: () => organizationService.list(params),
  });
}

export function useCreateOrganization() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: organizationService.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Organization created");
    },
    onError: () => toast.error("Failed to create organization"),
  });
}

export function useUpdateOrganization() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<any> }) =>
      organizationService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Organization updated");
    },
    onError: () => toast.error("Failed to update organization"),
  });
}

export function useDeleteOrganization() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => organizationService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Organization deleted");
    },
    onError: () => toast.error("Failed to delete organization"),
  });
}
