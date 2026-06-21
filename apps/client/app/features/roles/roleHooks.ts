import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { roleService } from "./roleService";
import { toast } from "@casbin/ui";

const KEY = ["roles"];

export function useRoles(params?: { page?: number; limit?: number }) {
  return useQuery({
    queryKey: [...KEY, params],
    queryFn: () => roleService.list(params),
  });
}

export function useCreateRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: roleService.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Role created");
    },
    onError: () => toast.error("Failed to create role"),
  });
}

export function useUpdateRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<any> }) =>
      roleService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Role updated");
    },
    onError: () => toast.error("Failed to update role"),
  });
}

export function useDeleteRole() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => roleService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Role deleted");
    },
    onError: () => toast.error("Failed to delete role"),
  });
}
