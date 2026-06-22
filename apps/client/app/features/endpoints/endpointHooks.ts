import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { endpointService } from "./endpointService";
import { toast } from "@casbin/ui";

const KEY = ["endpoints"];

export function useEndpoints(params?: {
  page?: number;
  limit?: number;
  resource_id?: string;
}) {
  return useQuery({
    queryKey: [...KEY, params],
    queryFn: () => endpointService.list(params),
  });
}

export function useCreateEndpoint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: endpointService.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Endpoint created");
    },
    onError: () => toast.error("Failed to create endpoint"),
  });
}

export function useUpdateEndpoint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<any> }) =>
      endpointService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Endpoint updated");
    },
    onError: () => toast.error("Failed to update endpoint"),
  });
}

export function useDeleteEndpoint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => endpointService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: KEY });
      toast.success("Endpoint deleted");
    },
    onError: () => toast.error("Failed to delete endpoint"),
  });
}
