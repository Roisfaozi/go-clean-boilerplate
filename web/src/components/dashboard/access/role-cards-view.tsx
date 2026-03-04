"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { RoleDialog } from "~/components/dashboard/roles/role-dialog";
import { Icon } from "~/components/shared/icon";
import { useDensity } from "~/components/shared/providers/density-provider";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Skeleton } from "~/components/ui/skeleton";
import { accessApi } from "~/lib/api/access";
import { rolesApi, type Role } from "~/lib/api/roles";

interface RoleCardsViewProps {
  onRoleClick?: (role: Role) => void;
}

interface RoleCardData extends Role {
  memberCount: number;
  resourceCount: number;
}

export function RoleCardsView({ onRoleClick }: RoleCardsViewProps) {
  const { density } = useDensity();
  const isCompact = density === "compact";
  const [roles, setRoles] = useState<RoleCardData[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [rolesResp, resourceResp] = await Promise.all([
        rolesApi.getAll(),
        accessApi.getResourceAggregation(),
      ]);

      const roleList = rolesResp.data ?? [];
      const resources = resourceResp.data?.resources ?? [];

      const enriched: RoleCardData[] = roleList.map((role: Role) => {
        const resourceCount = resources.filter((r) => {
          const crud = r.role_permissions[role.name];
          return (
            crud && (crud.create || crud.read || crud.update || crud.delete)
          );
        }).length;

        return {
          ...role,
          memberCount: 0,
          resourceCount,
        };
      });

      setRoles(enriched);
    } catch {
      toast.error("Failed to load roles");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (isLoading) {
    return (
      <div
        className={`grid grid-cols-1 ${isCompact ? "gap-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4" : "gap-4 sm:grid-cols-2 lg:grid-cols-3"}`}
      >
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton
            key={i}
            className={`${isCompact ? "h-36" : "h-44"} w-full rounded-xl`}
          />
        ))}
      </div>
    );
  }

  return (
    <>
      <div
        className={`grid grid-cols-1 ${isCompact ? "gap-2 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4" : "gap-4 sm:grid-cols-2 lg:grid-cols-3"}`}
      >
        {roles.map((role) => {
          const cleanName = role.name.replace("role:", "");
          const isSystem =
            role.name.startsWith("role:super") ||
            role.name === "role:admin" ||
            role.name === "role:user";

          return (
            <div
              key={role.id}
              className={`group relative flex flex-col justify-between rounded-xl border ${isCompact ? "p-3" : "p-5"} hover:border-primary/30 transition-all hover:shadow-md`}
            >
              <div>
                <div
                  className={`flex items-center gap-2 ${isCompact ? "mb-1.5" : "mb-3"}`}
                >
                  <div className="bg-primary/10 text-primary rounded-md p-1.5">
                    <Icon
                      name="Shield"
                      className={`${isCompact ? "h-3.5 w-3.5" : "h-4 w-4"}`}
                    />
                  </div>
                  <h3 className="font-semibold">{cleanName}</h3>
                  {isSystem && (
                    <Badge
                      variant="secondary"
                      className="px-1.5 py-0 text-[10px]"
                    >
                      System
                    </Badge>
                  )}
                </div>
                <p
                  className={`text-muted-foreground line-clamp-2 text-sm ${isCompact ? "mb-2 text-xs" : "mb-4"}`}
                >
                  {role.description || "No description"}
                </p>
              </div>

              <div>
                <div
                  className={`text-muted-foreground flex items-center gap-4 text-xs ${isCompact ? "mb-1.5" : "mb-3"}`}
                >
                  <div className="flex items-center gap-1">
                    <Icon name="Users" className="h-3.5 w-3.5" />
                    <span>{role.memberCount} members</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Icon name="Lock" className="h-3.5 w-3.5" />
                    <span>{role.resourceCount} resources</span>
                  </div>
                </div>

                <Button
                  variant="ghost"
                  size="sm"
                  className={`text-primary w-full justify-start px-0 ${isCompact ? "h-6 text-xs" : ""}`}
                  onClick={() => onRoleClick?.(role)}
                >
                  Edit →
                </Button>
              </div>
            </div>
          );
        })}

        <button
          type="button"
          onClick={() => setDialogOpen(true)}
          className={`flex flex-col items-center justify-center rounded-xl border-2 border-dashed ${isCompact ? "min-h-[140px] p-3" : "min-h-[176px] p-5"} hover:border-primary/40 hover:bg-muted/30 text-center transition-colors`}
        >
          <div className="bg-muted mb-2 rounded-full p-2">
            <Icon name="Plus" className="text-muted-foreground h-5 w-5" />
          </div>
          <span className="text-sm font-medium">Create Role</span>
          <span className="text-muted-foreground mt-1 text-xs">
            Add a new role to the system
          </span>
        </button>
      </div>

      <RoleDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        onSuccess={fetchData}
      />
    </>
  );
}
