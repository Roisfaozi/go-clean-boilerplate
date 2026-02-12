"use client";

import { PermissionMatrixProvider } from "~/app/[locale]/dashboard/access/_components/permission-matrix-context";
import { MatrixGrid } from "~/app/[locale]/dashboard/access/_components/matrix-grid";
import { MatrixDialog } from "~/app/[locale]/dashboard/access/_components/matrix-dialog";
import { Icon } from "~/components/shared/icon";
import { Role } from "~/lib/api/roles";

interface PermissionMatrixViewProps {
  onRoleClick?: (role: Role) => void;
}

export function PermissionMatrixView({
  onRoleClick,
}: PermissionMatrixViewProps) {
  return (
    <PermissionMatrixProvider>
      <div className="space-y-4">
        <div className="bg-primary/5 border-primary/10 flex items-center gap-3 rounded-lg border p-4">
          <div className="bg-primary/10 rounded-full p-2">
            <Icon name="Info" className="text-primary h-4 w-4" />
          </div>
          <p className="text-muted-foreground text-xs leading-relaxed">
            The Permission Matrix provides a high-level overview of CRUD access across all system resources. 
            Click a cell to modify granular permissions for a specific role and resource pair.
          </p>
        </div>
        
        <MatrixGrid />
        <MatrixDialog />

        <div className="text-muted-foreground flex items-center gap-6 text-[10px] uppercase font-bold tracking-widest">
          <div className="flex items-center gap-2">
            <div className="bg-primary h-3 w-3 rounded-sm shadow-sm" />
            <span>Active Access</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="bg-muted-foreground/15 h-3 w-3 rounded-sm" />
            <span>No Permission</span>
          </div>
          <div className="ml-auto flex gap-4">
            <span>C: Create</span>
            <span>R: Read</span>
            <span>U: Update</span>
            <span>D: Delete</span>
          </div>
        </div>
      </div>
    </PermissionMatrixProvider>
  );
}
