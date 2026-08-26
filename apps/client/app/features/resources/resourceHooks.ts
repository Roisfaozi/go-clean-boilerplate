import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { resourceService } from "./resourceService";
import { toast } from "@casbin/ui";

const KEY = ["resources"];

export function useResources() {
	return useQuery({ queryKey: KEY, queryFn: () => resourceService.list() });
}

export function useCreateResource() {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: resourceService.create,
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: KEY });
			toast.success("Resource created");
		},
		onError: () => toast.error("Failed to create resource"),
	});
}

export function useUpdateResource() {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: ({ id, data }: { id: string; data: Partial<any> }) =>
			resourceService.update(id, data),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: KEY });
			toast.success("Resource updated");
		},
		onError: () => toast.error("Failed to update resource"),
	});
}

export function useDeleteResource() {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: (id: string) => resourceService.delete(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: KEY });
			toast.success("Resource deleted");
		},
		onError: () => toast.error("Failed to delete resource"),
	});
}
