"use client";

import { useRoles } from "./roles-context";
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

export function RolesModals() {
  const {
    isDialogOpen,
    setIsDialogOpen,
    selectedRole,
    fetchRoles,
    isSheetOpen,
    setIsSheetOpen,
    isAlertOpen,
    setIsAlertOpen,
    confirmDelete,
  } = useRoles();

  return (
    <>
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
    </>
  );
}
