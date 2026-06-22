import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { userService } from "./userService";
import { toast } from "@casbin/ui";
import type { User } from "@/lib/api/schemas";

const USERS_KEY = ["users"];

export function useUsers(params?: {
  page?: number;
  limit?: number;
  search?: string;
}) {
  return useQuery({
    queryKey: [...USERS_KEY, params],
    queryFn: () => userService.list(params),
  });
}

export function useUser(id: string) {
  return useQuery({
    queryKey: [...USERS_KEY, id],
    queryFn: () => userService.get(id),
    enabled: !!id,
  });
}

export function useCreateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: userService.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: USERS_KEY });
      toast.success("User created");
    },
    onError: () => toast.error("Failed to create user"),
  });
}

export function useUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<User> }) =>
      userService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: USERS_KEY });
      toast.success("User updated");
    },
    onError: () => toast.error("Failed to update user"),
  });
}

export function useDeleteUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => userService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: USERS_KEY });
      toast.success("User deleted");
    },
    onError: () => toast.error("Failed to delete user"),
  });
}
