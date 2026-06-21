import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { accessRightService } from "./accessRightService";
import { toast } from "@casbin/ui";

const KEY = ["access-rights"];

export function useAccessRights(params?: { page?: number; limit?: number }) {
  return useQuery({
    queryKey: [...KEY, params],
    queryFn: () => accessRightService.list(params),
  });
}

export function useCreateAccessRight() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: accessRightService.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Access right created");
    },
    onError: () => toast.error("Failed to create access right"),
  });
}

export function useUpdateAccessRight() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<any> }) =>
      accessRightService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Access right updated");
    },
    onError: () => toast.error("Failed to update access right"),
  });
}

export function useDeleteAccessRight() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => accessRightService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Access right deleted");
    },
    onError: () => toast.error("Failed to delete access right"),
  });
}
