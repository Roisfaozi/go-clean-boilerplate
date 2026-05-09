import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { permissionService } from './permissionService';
import { toast } from '@casbin/ui';

const KEY = ['permissions'];

export function usePermissions(params?: { page?: number; limit?: number; role_id?: string }) {
  return useQuery({ queryKey: [...KEY, params], queryFn: () => permissionService.list(params) });
}

export function useCreatePermission() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: permissionService.create,
    onSuccess: () => { qc.invalidateQueries({ queryKey: KEY }); toast.success('Permission created'); },
    onError: () => toast.error('Failed to create permission'),
  });
}

export function useUpdatePermission() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<any> }) => permissionService.update(id, data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: KEY }); toast.success('Permission updated'); },
    onError: () => toast.error('Failed to update permission'),
  });
}

export function useDeletePermission() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => permissionService.delete(id),
    onSuccess: () => { qc.invalidateQueries({ queryKey: KEY }); toast.success('Permission deleted'); },
    onError: () => toast.error('Failed to delete permission'),
  });
}
