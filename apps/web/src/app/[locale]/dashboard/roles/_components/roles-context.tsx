"use client";

import { createContext, useContext, useState, useCallback, useEffect, ReactNode } from "react";
import { rolesApi, Role } from "~/lib/api/roles";
import { toast } from "sonner";

interface RolesContextType {
  roles: Role[];
  isLoading: boolean;
  isDialogOpen: boolean;
  setIsDialogOpen: (open: boolean) => void;
  isSheetOpen: boolean;
  setIsSheetOpen: (open: boolean) => void;
  isAlertOpen: boolean;
  setIsAlertOpen: (open: boolean) => void;
  selectedRole?: Role;
  setSelectedRole: (role?: Role) => void;
  fetchRoles: () => Promise<void>;
  handleCreate: () => void;
  handleEdit: (role: Role) => void;
  handleDetail: (role: Role) => void;
  handleDelete: (role: Role) => void;
  confirmDelete: () => Promise<void>;
}

const RolesContext = createContext<RolesContextType | undefined>(undefined);

export function RolesProvider({ children }: { children: ReactNode }) {
  const [roles, setRoles] = useState<Role[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [isAlertOpen, setIsAlertOpen] = useState(false);
  const [selectedRole, setSelectedRole] = useState<Role | undefined>(undefined);

  const fetchRoles = useCallback(async () => {
    setIsLoading(true);
    try {
      const response = await rolesApi.getAll();
      if (response && response.data) {
        setRoles(response.data);
      } else {
        setRoles([]);
      }
    } catch (error) {
      console.error("Failed to fetch roles:", error);
      toast.error("Failed to fetch roles");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchRoles();
  }, [fetchRoles]);

  const handleCreate = useCallback(() => {
    setSelectedRole(undefined);
    setIsDialogOpen(true);
  }, []);

  const handleEdit = useCallback((role: Role) => {
    setSelectedRole(role);
    setIsDialogOpen(true);
  }, []);

  const handleDetail = useCallback((role: Role) => {
    setSelectedRole(role);
    setIsSheetOpen(true);
  }, []);

  const handleDelete = useCallback((role: Role) => {
    setSelectedRole(role);
    setIsAlertOpen(true);
  }, []);

  const confirmDelete = useCallback(async () => {
    if (!selectedRole) return;
    try {
      await rolesApi.delete(selectedRole.id);
      toast.success("Role deleted successfully");
      await fetchRoles();
    } catch (error) {
      toast.error("Failed to delete role");
    } finally {
      setIsAlertOpen(false);
      setSelectedRole(undefined);
    }
  }, [selectedRole, fetchRoles]);

  return (
    <RolesContext.Provider
      value={{
        roles,
        isLoading,
        isDialogOpen,
        setIsDialogOpen,
        isSheetOpen,
        setIsSheetOpen,
        isAlertOpen,
        setIsAlertOpen,
        selectedRole,
        setSelectedRole,
        fetchRoles,
        handleCreate,
        handleEdit,
        handleDetail,
        handleDelete,
        confirmDelete,
      }}
    >
      {children}
    </RolesContext.Provider>
  );
}

export function useRoles() {
  const context = useContext(RolesContext);
  if (context === undefined) {
    throw new Error("useRoles must be used within a RolesProvider");
  }
  return context;
}
