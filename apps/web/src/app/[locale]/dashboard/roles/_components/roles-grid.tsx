"use client";

import { useRoles } from "./roles-context";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { Role } from "~/lib/api/roles";
import { memo } from "react";
import { EmptyState } from "~/components/shared/empty-state";
import { CardGridSkeleton } from "~/components/shared/skeletons";

export function RolesGrid() {
  const { roles, isLoading, handleCreate } = useRoles();

  if (isLoading && roles.length === 0) {
    return <CardGridSkeleton count={3} />;
  }

  if (roles.length === 0) {
    return (
      <EmptyState
        case="roles"
        action={{
          label: "Create Your First Role",
          onClick: handleCreate,
          icon: "Plus",
        }}
      />
    );
  }

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <CreateRoleCard onClick={handleCreate} />
      {roles.map((role) => (
        <RoleCard key={role.id} role={role} />
      ))}
    </div>
  );
}

function CreateRoleCard({ onClick }: { onClick: () => void }) {
  return (
    <Card
      role="button"
      onClick={onClick}
      className="hover:bg-accent flex flex-col items-center justify-center gap-y-2.5 border-dashed p-8 text-center transition-colors"
    >
      <div className="bg-primary/10 text-primary rounded-full p-3">
        <Icon name="Plus" className="h-6 w-6" />
      </div>
      <p className="text-lg font-semibold">Create Role</p>
      <p className="text-muted-foreground text-xs font-bold tracking-widest uppercase">
        New RBAC Policy
      </p>
    </Card>
  );
}

const RoleCard = memo(function RoleCard({ role }: { role: Role }) {
  const { handleDetail, handleEdit, handleDelete } = useRoles();

  return (
    <Card className="group hover:border-primary/50 transition-colors">
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="bg-primary/10 text-primary rounded-md p-2">
              <Icon name="Shield" className="h-5 w-5" />
            </div>
            <CardTitle className="text-lg">{role.name}</CardTitle>
          </div>
          {role.name.startsWith("role:") ? (
            <Badge
              variant="secondary"
              className="bg-primary/5 text-primary border-primary/10"
            >
              System
            </Badge>
          ) : (
            <Badge variant="outline">Custom</Badge>
          )}
        </div>
        <CardDescription className="line-clamp-2 min-h-[2.5rem]">
          {role.description || "No description provided."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="text-muted-foreground space-y-2 text-sm">
          <div className="flex items-center gap-2">
            <Icon name="Users" className="h-4 w-4" />
            <span>Manage members</span>
          </div>
          <div className="flex items-center gap-2">
            <Icon name="Lock" className="h-4 w-4" />
            <span>View permissions</span>
          </div>
        </div>
      </CardContent>
      <CardFooter className="bg-muted/20 flex justify-end gap-2 border-t p-4">
        <Button variant="outline" size="sm" onClick={() => handleDetail(role)}>
          Manage
        </Button>
        <Button variant="ghost" size="sm" onClick={() => handleEdit(role)}>
          Edit
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="text-destructive hover:text-destructive hover:bg-destructive/10"
          onClick={() => handleDelete(role)}
          disabled={role.name === "role:superadmin"}
        >
          <Icon name="Trash2" className="h-4 w-4" />
        </Button>
      </CardFooter>
    </Card>
  );
});
