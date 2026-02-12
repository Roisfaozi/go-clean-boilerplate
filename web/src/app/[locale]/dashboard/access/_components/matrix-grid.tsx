"use client";

import { memo, useState } from "react";
import { usePermissionMatrix } from "./permission-matrix-context";
import { Icon } from "~/components/shared/icon";
import { Badge } from "~/components/ui/badge";
import { useDensity } from "~/components/shared/providers/density-provider";
import { Skeleton } from "~/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/components/ui/tooltip";
import { ResourceCRUD } from "~/lib/api/access";

export function MatrixGrid() {
  const { resources, roles, isLoading } = usePermissionMatrix();
  const [hoveredRow, setHoveredRow] = useState<string | null>(null);
  const { density } = useDensity();
  const isCompact = density === "compact";

  if (isLoading && resources.length === 0) {
    return (
      <div className="space-y-3">
        <Skeleton className="h-10 w-full" />
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-14 w-full" />
        ))}
      </div>
    );
  }

  return (
    <div className="overflow-x-auto rounded-lg border bg-card">
      <table className="w-full border-collapse text-sm">
        <thead>
          <tr className="bg-muted/50">
            <th
              className={`text-muted-foreground bg-muted/50 sticky left-0 z-10 text-left text-xs font-semibold tracking-wider uppercase border-r ${
                isCompact ? "px-2 py-2" : "px-4 py-3"
              }`}
            >
              Role / Resource
            </th>
            {resources.map((resource) => (
              <th
                key={resource.name}
                className={`${isCompact ? "px-2 py-2" : "px-4 py-3"} text-center border-r last:border-r-0`}
              >
                <div className="flex flex-col items-center gap-0.5 min-w-[80px]">
                  <code className="font-mono text-[11px] font-bold text-primary">
                    /{resource.name.toLowerCase()}
                  </code>
                  <span className="text-muted-foreground text-[9px]">
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
                className={`bg-card sticky left-0 z-10 border-r font-medium ${
                  isCompact ? "px-2 py-2" : "px-4 py-3"
                }`}
              >
                <div className="flex items-center justify-between gap-4">
                  <span className="truncate">{role.name.replace("role:", "")}</span>
                  {role.name === "role:superadmin" && (
                    <Badge variant="secondary" className="h-4 px-1 text-[8px] uppercase">Root</Badge>
                  )}
                </div>
              </td>
              {resources.map((resource) => (
                <MatrixCell
                  key={`${role.name}-${resource.name}`}
                  roleName={role.name}
                  resourceName={resource.name}
                  crud={resource.role_permissions[role.name] || { create: false, read: false, update: false, delete: false }}
                  isCompact={isCompact}
                />
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

const CRUD_LABELS = ["C", "R", "U", "D"] as const;

const MatrixCell = memo(function MatrixCell({
  roleName,
  resourceName,
  crud,
  isCompact,
}: {
  roleName: string;
  resourceName: string;
  crud: ResourceCRUD;
  isCompact: boolean;
}) {
  const { openDialog } = usePermissionMatrix();
  const flags = [crud.create, crud.read, crud.update, crud.delete];

  return (
    <td className={`border-r last:border-r-0 ${isCompact ? "px-1 py-1" : "px-2 py-2"} text-center`}>
      <div className="flex items-center justify-center">
        <TooltipProvider delayDuration={100}>
          <Tooltip>
            <TooltipTrigger asChild>
              <button
                type="button"
                onClick={() => openDialog(roleName, resourceName, crud)}
                className="group/cell hover:ring-primary/40 flex cursor-pointer items-center gap-[2px] rounded-md p-1.5 transition-all hover:scale-110 hover:ring-2 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={roleName === "role:superadmin"}
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
            <TooltipContent side="top" className="text-[10px]">
              <div className="font-semibold mb-1 uppercase tracking-wider">{resourceName}</div>
              <div className="flex gap-2">
                {CRUD_LABELS.map((label, i) => (
                  <div key={label} className={flags[i] ? "text-primary font-bold" : "text-muted-foreground opacity-40"}>
                    {label}
                  </div>
                ))}
              </div>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    </td>
  );
});
