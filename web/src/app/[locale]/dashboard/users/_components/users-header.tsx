"use client";

import { useUsers } from "./users-context";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { useMounted } from "~/hooks/use-mounted";

export function UsersHeader() {
  const { canCreate, handleCreate } = useUsers();
  const isMounted = useMounted();

  return (
    <div className="flex items-center justify-between">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Users</h2>
        <p className="text-muted-foreground">
          Manage your team members and their account permissions here.
        </p>
      </div>
      <div className="flex items-center space-x-2">
        {isMounted && canCreate && (
          <Button onClick={handleCreate}>
            <Icon name="UserPlus" className="mr-2 h-4 w-4" />
            Add User
          </Button>
        )}
      </div>
    </div>
  );
}
