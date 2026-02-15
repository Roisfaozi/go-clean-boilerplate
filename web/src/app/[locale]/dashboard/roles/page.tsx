"use client";

import { useState, useEffect, useCallback } from "react";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { rolesApi, Role } from "~/lib/api/roles";
import { toast } from "sonner";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Badge } from "~/components/ui/badge";
import { RoleDialog } from "~/components/dashboard/roles/role-dialog";
import { RoleDetailSheet } from "~/components/dashboard/roles/role-detail-sheet";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "~/components/ui/alert-dialog";

export default function RolesPage() {
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

  const handleCreate = () => {
    setSelectedRole(undefined);
    setIsDialogOpen(true);
  };

  const handleEdit = (role: Role) => {
    setSelectedRole(role);
    setIsDialogOpen(true);
  };

  const handleDetail = (role: Role) => {
    setSelectedRole(role);
    setIsSheetOpen(true);
  };

  const handleDelete = (role: Role) => {
    setSelectedRole(role);
    setIsAlertOpen(true);
  };

  const confirmDelete = async () => {
    if (!selectedRole) return;
    try {
      await rolesApi.delete(selectedRole.id);
      toast.success("Role deleted successfully");
      fetchRoles();
    } catch (error) {
      toast.error("Failed to delete role");
    } finally {
      setIsAlertOpen(false);
      setSelectedRole(undefined);
    }
  };

  return (
    <div className="space-y-6">
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

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {isLoading
          ? Array.from({ length: 3 }).map((_, i) => (
              <Card key={i} className="animate-pulse">
                <CardHeader className="bg-muted/50 h-24" />
                <CardContent className="h-32" />
              </Card>
            ))
          : roles.map((role) => (
              <Card
                key={role.id}
                className="group hover:border-primary/50 transition-colors"
              >
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
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleDetail(role)}
                  >
                    Manage
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleEdit(role)}
                  >
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
            ))}
      </div>

      <RoleDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        role={selectedRole}
        onSuccess={fetchRoles}
      />

      <RoleDetailSheet
        open={isSheetOpen}
        onOpenChange={setIsSheetOpen}
        role={selectedRole}
      />

      <AlertDialog open={isAlertOpen} onOpenChange={setIsAlertOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This action cannot be undone. This will permanently delete the
              role and remove all associated permissions.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
