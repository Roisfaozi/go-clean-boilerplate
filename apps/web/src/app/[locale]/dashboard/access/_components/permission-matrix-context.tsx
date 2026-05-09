"use client";

import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  ReactNode,
} from "react";
import { toast } from "sonner";
import { accessApi, ResourceCRUD, ResourcePermission } from "~/lib/api/access";
import { rolesApi, Role } from "~/lib/api/roles";

interface PermissionMatrixContextType {
  resources: ResourcePermission[];
  roles: Role[];
  isLoading: boolean;
  isProcessing: boolean;
  fetchData: () => Promise<void>;
  updatePermissions: (
    roleName: string,
    resourceName: string,
    newCrud: ResourceCRUD,
    oldCrud: ResourceCRUD
  ) => Promise<void>;

  // Dialog state
  dialog: {
    open: boolean;
    resource: string;
    role: string;
    crud: ResourceCRUD;
  };
  openDialog: (role: string, resource: string, crud: ResourceCRUD) => void;
  closeDialog: () => void;
}

const PermissionMatrixContext = createContext<
  PermissionMatrixContextType | undefined
>(undefined);

export function PermissionMatrixProvider({
  children,
}: {
  children: ReactNode;
}) {
  const [resources, setResources] = useState<ResourcePermission[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isProcessing, setIsProcessing] = useState(false);

  const [dialog, setDialog] = useState({
    open: false,
    resource: "",
    role: "",
    crud: {
      create: false,
      read: false,
      update: false,
      delete: false,
    } as ResourceCRUD,
  });

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [resourceResp, rolesResp] = await Promise.all([
        accessApi.getResourceAggregation(),
        rolesApi.getAll(),
      ]);

      if (resourceResp.data?.resources) {
        setResources(resourceResp.data.resources);
      }
      if (rolesResp.data) {
        setRoles(rolesResp.data);
      }
    } catch (error) {
      toast.error("Failed to load permission matrix");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const updatePermissions = useCallback(
    async (
      roleName: string,
      resourceName: string,
      newCrud: ResourceCRUD,
      oldCrud: ResourceCRUD
    ) => {
      const resource = resources.find((r) => r.name === resourceName);
      if (!resource) return;

      setIsProcessing(true);
      const basePath = resource.base_path;
      const methodMap: Record<keyof ResourceCRUD, string> = {
        create: "POST",
        read: "GET",
        update: "PUT",
        delete: "DELETE",
      };

      const CRUD_KEYS: (keyof ResourceCRUD)[] = [
        "create",
        "read",
        "update",
        "delete",
      ];
      const promises: Promise<any>[] = [];

      for (const key of CRUD_KEYS) {
        if (oldCrud[key] !== newCrud[key]) {
          if (newCrud[key]) {
            promises.push(
              accessApi.grantPermission(roleName, basePath, methodMap[key])
            );
            promises.push(
              accessApi.grantPermission(
                roleName,
                `${basePath}/*`,
                methodMap[key]
              )
            );
          } else {
            promises.push(
              accessApi.revokePermission(roleName, basePath, methodMap[key])
            );
            promises.push(
              accessApi.revokePermission(
                roleName,
                `${basePath}/*`,
                methodMap[key]
              )
            );
          }
        }
      }

      try {
        await Promise.all(promises);
        toast.success(
          `Permissions updated for ${roleName.replace("role:", "")}`
        );
        await fetchData();
      } catch (error) {
        toast.error("Failed to update some permissions");
      } finally {
        setIsProcessing(false);
      }
    },
    [resources, fetchData]
  );

  const openDialog = useCallback(
    (role: string, resource: string, crud: ResourceCRUD) => {
      setDialog({ open: true, role, resource, crud });
    },
    []
  );

  const closeDialog = useCallback(() => {
    setDialog((prev) => ({ ...prev, open: false }));
  }, []);

  return (
    <PermissionMatrixContext.Provider
      value={{
        resources,
        roles,
        isLoading,
        isProcessing,
        fetchData,
        updatePermissions,
        dialog,
        openDialog,
        closeDialog,
      }}
    >
      {children}
    </PermissionMatrixContext.Provider>
  );
}

export function usePermissionMatrix() {
  const context = useContext(PermissionMatrixContext);
  if (context === undefined) {
    throw new Error(
      "usePermissionMatrix must be used within a PermissionMatrixProvider"
    );
  }
  return context;
}
