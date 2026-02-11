"use client";

import {
  createContext,
  useContext,
  useState,
  useCallback,
  ReactNode,
  useEffect,
} from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import useSWR, { KeyedMutator } from "swr";
import { User, usersApi } from "~/lib/api/users";
import { toast } from "sonner";
import { usePermission } from "~/hooks/use-permission";

interface UsersContextType {
  users: User[];
  total: number;
  page: number;
  limit: number;
  searchTerm: string;
  isLoading: boolean;
  error: any;
  mutate: KeyedMutator<any>;
  handleSearch: (term: string) => void;
  handlePageChange: (newPage: number) => void;
  canCreate: boolean;
  canDelete: boolean;
  canUpdate: boolean;

  // Modal states
  isDialogOpen: boolean;
  setIsDialogOpen: (open: boolean) => void;
  isAlertOpen: boolean;
  setIsAlertOpen: (open: boolean) => void;
  selectedUser?: User;
  setSelectedUser: (user?: User) => void;

  handleCreate: () => void;
  handleEdit: (user: User) => void;
  handleDelete: (user: User) => void;
  confirmDelete: () => Promise<void>;
}

const UsersContext = createContext<UsersContextType | undefined>(undefined);

export function UsersProvider({ children }: { children: ReactNode }) {
  const searchParams = useSearchParams();
  const pathname = usePathname();
  const { replace } = useRouter();

  const page = Number(searchParams.get("page")) || 1;
  const limit = Number(searchParams.get("limit")) || 10;
  const searchTerm = searchParams.get("search") || "";

  const {
    data: response,
    error,
    isLoading,
    mutate,
  } = useSWR(["/api/v1/users", page, limit, searchTerm], ([_, p, l, s]) =>
    usersApi.getAll(p, l, s)
  );

  const users = response?.data || [];
  const total = response?.paging.total || 0;

  const canCreate = usePermission("/api/v1/users", "POST");
  const canDelete = usePermission("/api/v1/users/:id", "DELETE");
  const canUpdate = usePermission("/api/v1/users/:id", "PUT");

  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isAlertOpen, setIsAlertOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | undefined>(undefined);

  const handleSearch = useCallback(
    (term: string) => {
      const params = new URLSearchParams(searchParams);
      if (term) {
        params.set("search", term);
      } else {
        params.delete("search");
      }
      params.set("page", "1");
      replace(`${pathname}?${params.toString()}`);
    },
    [searchParams, pathname, replace]
  );

  const handlePageChange = useCallback(
    (newPage: number) => {
      const params = new URLSearchParams(searchParams);
      params.set("page", newPage.toString());
      replace(`${pathname}?${params.toString()}`);
    },
    [searchParams, pathname, replace]
  );

  const handleCreate = useCallback(() => {
    setSelectedUser(undefined);
    setIsDialogOpen(true);
  }, []);

  const handleEdit = useCallback((user: User) => {
    setSelectedUser(user);
    setIsDialogOpen(true);
  }, []);

  const handleDelete = useCallback((user: User) => {
    setSelectedUser(user);
    setIsAlertOpen(true);
  }, []);

  const confirmDelete = useCallback(async () => {
    if (!selectedUser) return;
    try {
      await usersApi.delete(selectedUser.id);
      toast.success("User deleted successfully");
      mutate();
    } catch (error) {
      toast.error("Failed to delete user");
    } finally {
      setIsAlertOpen(false);
      setSelectedUser(undefined);
    }
  }, [selectedUser, mutate]);

  return (
    <UsersContext.Provider
      value={{
        users,
        total,
        page,
        limit,
        searchTerm,
        isLoading,
        error,
        mutate,
        handleSearch,
        handlePageChange,
        canCreate,
        canDelete,
        canUpdate,
        isDialogOpen,
        setIsDialogOpen,
        isAlertOpen,
        setIsAlertOpen,
        selectedUser,
        setSelectedUser,
        handleCreate,
        handleEdit,
        handleDelete,
        confirmDelete,
      }}
    >
      {children}
    </UsersContext.Provider>
  );
}

export function useUsers() {
  const context = useContext(UsersContext);
  if (context === undefined) {
    throw new Error("useUsers must be used within a UsersProvider");
  }
  return context;
}
