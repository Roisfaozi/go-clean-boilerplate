"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import { Checkbox } from "~/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { type ResourceCRUD } from "~/lib/api/access";

interface CRUDPermissionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  resourceName: string;
  roleName: string;
  currentPermissions: ResourceCRUD;
  onApply: (permissions: ResourceCRUD) => Promise<void>;
}

const CRUD_ITEMS = [
  {
    key: "read" as const,
    label: "Read",
    icon: "Eye",
    description: "View and list resources",
  },
  {
    key: "create" as const,
    label: "Create",
    icon: "Plus",
    description: "Add new resources to the system",
  },
  {
    key: "update" as const,
    label: "Update",
    icon: "Pencil",
    description: "Modify existing resources",
  },
  {
    key: "delete" as const,
    label: "Delete",
    icon: "Trash2",
    description: "Remove resources permanently",
  },
];

export function CRUDPermissionDialog({
  open,
  onOpenChange,
  resourceName,
  roleName,
  currentPermissions,
  onApply,
}: CRUDPermissionDialogProps) {
  const [permissions, setPermissions] =
    useState<ResourceCRUD>(currentPermissions);
  const [isApplying, setIsApplying] = useState(false);

  const handleToggle = (key: keyof ResourceCRUD) => {
    setPermissions((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const handleApply = async () => {
    setIsApplying(true);
    try {
      await onApply(permissions);
      onOpenChange(false);
    } catch {
      toast.error("Failed to update permissions");
    } finally {
      setIsApplying(false);
    }
  };

  const hasChanges =
    permissions.create !== currentPermissions.create ||
    permissions.read !== currentPermissions.read ||
    permissions.update !== currentPermissions.update ||
    permissions.delete !== currentPermissions.delete;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[420px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <code className="bg-muted rounded px-2 py-0.5 font-mono text-sm">
              {resourceName}
            </code>
          </DialogTitle>
          <DialogDescription>
            Permissions for{" "}
            <span className="text-foreground font-medium">{roleName}</span>
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-1 py-2">
          {CRUD_ITEMS.map((item) => (
            <label
              key={item.key}
              className="hover:bg-muted/50 flex cursor-pointer items-center gap-3 rounded-lg p-3 transition-colors"
            >
              <Checkbox
                checked={permissions[item.key]}
                onCheckedChange={() => handleToggle(item.key)}
                className="data-[state=checked]:bg-primary"
              />
              <div className="flex-1 space-y-0.5">
                <div className="flex items-center gap-2">
                  <Icon
                    name={item.icon as any}
                    className="text-muted-foreground h-3.5 w-3.5"
                  />
                  <span className="text-sm font-medium">{item.label}</span>
                </div>
                <p className="text-muted-foreground text-xs">
                  {item.description}
                </p>
              </div>
            </label>
          ))}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleApply} disabled={isApplying || !hasChanges}>
            {isApplying && (
              <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
            )}
            Apply
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
