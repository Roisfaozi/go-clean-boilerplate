"use client";

import { useRoles } from "./roles-context";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";

export function RolesHeader() {
  const { handleCreate } = useRoles();

  return (
    <div className="flex items-center justify-between">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Roles & Access</h2>
        <p className="text-muted-foreground">
          Manage system roles and their permissions.
        </p>
      </div>
      <Button onClick={handleCreate}>
        <Icon name="Plus" className="mr-2 h-4 w-4" />
        Create Role
      </Button>
    </div>
  );
}
