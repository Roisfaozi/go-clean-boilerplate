import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { permissionService } from "./permissionService";
import { toast } from "@casbin/ui";

const KEY = ["permissions"];

export function usePermissions(params?: {
	page?: number;
	limit?: number;
	role_id?: string;
}) {
	return useQuery({
		queryKey: [...KEY, params],
		queryFn: () => permissionService.list(params),
	});
}

export function useCreatePermission() {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: permissionService.create,
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: KEY });
			toast.success("Permission created");
		},
		onError: () => toast.error("Failed to create permission"),
	});
}

export function useUpdatePermission() {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: ({ id, data }: { id: string; data: Partial<any> }) =>
			permissionService.update(id, data),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: KEY });
			toast.success("Permission updated");
		},
		onError: () => toast.error("Failed to update permission"),
	});
}

export function useDeletePermission() {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: (id: string) => permissionService.delete(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: KEY });
			toast.success("Permission deleted");
		},
		onError: () => toast.error("Failed to delete permission"),
	});
}

export function useResourceAggregation() {
	return useQuery({
		queryKey: ["permissions", "aggregation"],
		queryFn: () => permissionService.getResources(),
	});
}

export function useRoleAccessRights(role: string, domain?: string) {
	return useQuery({
		queryKey: ["permissions", "role-access", role, domain],
		queryFn: () => permissionService.getRoleAccessRights(role, domain),
		enabled: !!role,
	});
}

export function useToggleAccessRight() {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: ({
			role,
			access_right_id,
			granted,
			domain,
		}: {
			role: string;
			access_right_id: string;
			granted: boolean;
			domain?: string;
		}) => {
			if (granted) {
				return permissionService.assignAccessRight({
					role,
					access_right_id,
					domain,
				});
			} else {
				return permissionService.revokeAccessRight({
					role,
					access_right_id,
					domain,
				});
			}
		},
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ["permissions"] });
			toast.success("Permission updated");
		},
		onError: () => toast.error("Failed to update permission"),
	});
}

export function useInheritanceTree() {
	return useQuery({
		queryKey: ["permissions", "inheritance-tree"],
		queryFn: () => permissionService.getInheritanceTree(),
	});
}

export function useAddInheritance() {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: permissionService.addInheritance,
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ["permissions", "inheritance-tree"] });
			toast.success("Role inheritance added");
		},
		onError: () => toast.error("Failed to add role inheritance"),
	});
}

export function useRemoveInheritance() {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: permissionService.removeInheritance,
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: ["permissions", "inheritance-tree"] });
			toast.success("Role inheritance removed");
		},
		onError: () => toast.error("Failed to remove role inheritance"),
	});
}
