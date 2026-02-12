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

export function RolesGrid() {
  const { roles, isLoading } = useRoles();

  if (isLoading && roles.length === 0) {
    return (
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader className="bg-muted/50 h-24" />
            <CardContent className="h-32" />
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      {roles.map((role) => (
        <RoleCard key={role.id} role={role} />
      ))}
    </div>
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
