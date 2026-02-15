"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { useDensity } from "~/components/shared/providers/density-provider";
import { Skeleton } from "~/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/components/ui/tooltip";
import {
  accessApi,
  type ResourceCRUD,
  type ResourcePermission,
} from "~/lib/api/access";
import { rolesApi, type Role } from "~/lib/api/roles";
import { CRUDPermissionDialog } from "./crud-permission-dialog";

interface PermissionMatrixViewProps {
  onRoleClick?: (role: Role) => void;
}

const CRUD_LABELS = ["C", "R", "U", "D"] as const;
const CRUD_KEYS: (keyof ResourceCRUD)[] = [
  "create",
  "read",
  "update",
  "delete",
];

function CRUDBlocks({
  crud,
  onClick,
  tooltip,
}: {
  crud: ResourceCRUD;
  onClick?: () => void;
  tooltip?: string;
}) {
  const flags = [crud.create, crud.read, crud.update, crud.delete];

  return (
    <TooltipProvider delayDuration={200}>
      <Tooltip>
        <TooltipTrigger asChild>
          <button
            type="button"
            onClick={onClick}
            className="group/cell hover:ring-primary/40 flex cursor-pointer items-center gap-[2px] rounded-md p-1.5 transition-all hover:scale-110 hover:ring-2"
          >
            {flags.map((enabled, i) => (
              <div
                key={CRUD_LABELS[i]}
                className={`h-5 w-2.5 rounded-[2px] transition-colors ${
                  enabled ? "bg-primary shadow-sm" : "bg-muted-foreground/15"
                }`}
              />
            ))}
          </button>
        </TooltipTrigger>
        <TooltipContent side="top" className="text-xs">
          {tooltip ??
            (flags
              .map((f, i) => (f ? CRUD_LABELS[i] : null))
              .filter(Boolean)
              .join(", ") ||
              "No permissions")}
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

export function PermissionMatrixView({
  onRoleClick,
}: PermissionMatrixViewProps) {
  const [resources, setResources] = useState<ResourcePermission[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [hoveredRow, setHoveredRow] = useState<string | null>(null);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [dialogResource, setDialogResource] = useState<string>("");
  const [dialogRole, setDialogRole] = useState<string>("");
  const [dialogCRUD, setDialogCRUD] = useState<ResourceCRUD>({
    create: false,
    read: false,
    update: false,
    delete: false,
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
    } catch {
      toast.error("Failed to load permission matrix");
    } finally {
      setIsLoading(false);
    }
  }, []);

  const { density } = useDensity();
  const isCompact = density === "compact";

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCellClick = (resourceName: string, roleName: string) => {
    const resource = resources.find((r) => r.name === resourceName);
    const crud = resource?.role_permissions[roleName] ?? {
      create: false,
      read: false,
      update: false,
      delete: false,
    };
    setDialogResource(resourceName);
    setDialogRole(roleName);
    setDialogCRUD(crud);
    setDialogOpen(true);
  };

  const handleApplyPermissions = async (newCrud: ResourceCRUD) => {
    const resource = resources.find((r) => r.name === dialogResource);
    if (!resource) return;

    const basePath = resource.base_path;
    const methodMap: Record<keyof ResourceCRUD, string> = {
      create: "POST",
      read: "GET",
      update: "PUT",
      delete: "DELETE",
    };

    const promises: Promise<any>[] = [];

    for (const key of CRUD_KEYS) {
      const oldVal = dialogCRUD[key];
      const newVal = newCrud[key];
      if (oldVal !== newVal) {
        if (newVal) {
          promises.push(
            accessApi.grantPermission(dialogRole, basePath, methodMap[key])
          );
          const wildcard = `${basePath}/*`;
          promises.push(
            accessApi.grantPermission(dialogRole, wildcard, methodMap[key])
          );
        } else {
          promises.push(
            accessApi.revokePermission(dialogRole, basePath, methodMap[key])
          );
          const wildcard = `${basePath}/*`;
          promises.push(
            accessApi.revokePermission(dialogRole, wildcard, methodMap[key])
          );
        }
      }
    }

    await Promise.all(promises);
    toast.success(`Permissions updated for ${dialogRole} on ${dialogResource}`);
    fetchData();
  };

  const getMemberCount = (roleName: string): string => {
    // Placeholder - in real app would need to fetch member counts
    return "";
  };

  if (isLoading) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-10 w-full" />
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-14 w-full" />
        ))}
      </div>
    );
  }

  if (resources.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border-2 border-dashed py-16">
        <Icon
          name="Table"
          className="text-muted-foreground/30 mb-3 h-10 w-10"
        />
        <p className="text-muted-foreground text-sm">
          No resources found. Add endpoints and access rights first.
        </p>
      </div>
    );
  }

  return (
    <>
      <div className="overflow-x-auto rounded-lg border">
        <table className="w-full border-collapse text-sm">
          <thead>
            <tr className="bg-muted/50">
              <th
                className={`text-muted-foreground bg-muted/50 sticky left-0 z-10 text-left text-xs font-medium tracking-wider uppercase ${isCompact ? "px-2 py-2" : "px-4 py-3"}`}
              >
                Role
              </th>
              {resources.map((resource) => (
                <th
                  key={resource.name}
                  className={`${isCompact ? "px-2 py-2" : "px-4 py-3"} text-center`}
                >
                  <div className="flex flex-col items-center gap-0.5">
                    <code className="font-mono text-xs font-medium">
                      /{resource.name.toLowerCase()}
                    </code>
                    <span className="text-muted-foreground text-[10px]">
                      {resource.endpoint_count} endpoints
                    </span>
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {roles.map((role) => (
              <tr
                key={role.name}
                className={`group border-t transition-colors ${
                  hoveredRow === role.name ? "bg-muted/30" : ""
                }`}
                onMouseEnter={() => setHoveredRow(role.name)}
                onMouseLeave={() => setHoveredRow(null)}
              >
                <td
                  className={`bg-background sticky left-0 z-10 ${isCompact ? "px-2 py-2" : "px-4 py-3"}`}
                >
                  <div className="flex items-center justify-between gap-4">
                    <button
                      type="button"
                      onClick={() => onRoleClick?.(role)}
                      className="flex flex-col items-start text-left hover:underline"
                    >
                      <span className="font-medium">
                        {role.name.replace("role:", "")}
                      </span>
                      {getMemberCount(role.name) && (
                        <span className="text-muted-foreground text-[11px]">
                          {getMemberCount(role.name)}
                        </span>
                      )}
                    </button>
                    <span
                      className={`text-primary text-xs opacity-0 transition-opacity ${
                        hoveredRow === role.name ? "opacity-100" : ""
                      }`}
                    >
                      Edit →
                    </span>
                  </div>
                </td>
                {resources.map((resource) => {
                  const crud = resource.role_permissions[role.name] ?? {
                    create: false,
                    read: false,
                    update: false,
                    delete: false,
                  };
                  return (
                    <td
                      key={`${role.name}-${resource.name}`}
                      className={`${isCompact ? "px-2 py-1" : "px-4 py-3"} text-center`}
                    >
                      <div className="flex items-center justify-center">
                        <CRUDBlocks
                          crud={crud}
                          onClick={() =>
                            handleCellClick(resource.name, role.name)
                          }
                          tooltip={`${resource.name} for ${role.name.replace("role:", "")}`}
                        />
                      </div>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="text-muted-foreground mt-4 flex items-center gap-6 text-xs">
        <div className="flex items-center gap-2">
          <div className="bg-primary h-3.5 w-2 rounded-[2px]" />
          <span>Enabled</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="bg-muted-foreground/15 h-3.5 w-2 rounded-[2px]" />
          <span>Disabled</span>
        </div>
        <div className="text-muted-foreground/60">
          C = Create &nbsp; R = Read &nbsp; U = Update &nbsp; D = Delete
        </div>
      </div>

      <CRUDPermissionDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        resourceName={dialogResource}
        roleName={dialogRole}
        currentPermissions={dialogCRUD}
        onApply={handleApplyPermissions}
      />
    </>
  );
}
