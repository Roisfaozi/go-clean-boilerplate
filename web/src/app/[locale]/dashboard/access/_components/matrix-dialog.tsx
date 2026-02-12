"use client";

import { usePermissionMatrix } from "./permission-matrix-context";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { Button } from "~/components/ui/button";
import { Checkbox } from "~/components/ui/checkbox";
import { Label } from "~/components/ui/label";
import { useState, useEffect } from "react";
import { ResourceCRUD } from "~/lib/api/access";
import { Icon } from "~/components/shared/icon";

export function MatrixDialog() {
  const { dialog, closeDialog, updatePermissions, isProcessing } = usePermissionMatrix();
  
  // Use a key to force re-mounting/resetting state when dialog target changes
  // instead of using useEffect which triggers cascading renders
  return (
    <MatrixDialogContent 
      key={`${dialog.role}-${dialog.resource}`}
      dialog={dialog}
      closeDialog={closeDialog}
      updatePermissions={updatePermissions}
      isProcessing={isProcessing}
    />
  );
}

function MatrixDialogContent({ 
  dialog, 
  closeDialog, 
  updatePermissions, 
  isProcessing 
}: { 
  dialog: any, 
  closeDialog: () => void, 
  updatePermissions: any, 
  isProcessing: boolean 
}) {
  const [localCRUD, setLocalCRUD] = useState<ResourceCRUD>(dialog.crud);

  const handleApply = async () => {
    await updatePermissions(dialog.role, dialog.resource, localCRUD, dialog.crud);
    closeDialog();
  };

  const toggle = (key: keyof ResourceCRUD) => {
    setLocalCRUD(prev => ({ ...prev, [key]: !prev[key] }));
  };

  return (
    <Dialog open={dialog.open} onOpenChange={(open) => !open && closeDialog()}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Icon name="Settings2" className="h-5 w-5 text-primary" />
            Edit Permissions
          </DialogTitle>
          <DialogDescription>
            Modify CRUD access for <span className="font-bold text-foreground">{dialog.role.replace("role:", "")}</span> on 
            <span className="font-mono bg-muted px-1 rounded ml-1">/{dialog.resource.toLowerCase()}</span>
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-6 py-6">
          {(["create", "read", "update", "delete"] as const).map((action) => (
            <div key={action} className="flex items-center justify-between space-x-2 rounded-lg border p-3 transition-colors hover:bg-muted/50">
              <div className="flex flex-col gap-0.5">
                <Label htmlFor={action} className="text-sm font-semibold capitalize">
                  {action}
                </Label>
                <p className="text-muted-foreground text-xs">
                  {getActionDescription(action)}
                </p>
              </div>
              <Checkbox
                id={action}
                checked={localCRUD[action]}
                onCheckedChange={() => toggle(action)}
                className="h-5 w-5"
              />
            </div>
          ))}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={closeDialog} disabled={isProcessing}>
            Cancel
          </Button>
          <Button onClick={handleApply} disabled={isProcessing}>
            {isProcessing && <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />}
            Apply Changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function getActionDescription(action: string) {
  switch (action) {
    case "create": return "Allow creating new resources";
    case "read": return "Allow viewing resource data";
    case "update": return "Allow modifying existing resources";
    case "delete": return "Allow removing resources";
    default: return "";
  }
}
