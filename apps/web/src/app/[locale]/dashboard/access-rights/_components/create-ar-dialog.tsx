"use client";

import { useState } from "react";
import { useAccessRights } from "./access-rights-context";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/components/ui/dialog";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";

export function CreateArDialog() {
  const { createAccessRight, isCreating } = useAccessRights();
  const [name, setName] = useState("");
  const [desc, setDesc] = useState("");
  const [isOpen, setIsOpen] = useState(false);

  const handleSubmit = async () => {
    if (!name) return;
    try {
      await createAccessRight(name, desc);
      setIsOpen(false);
      setName("");
      setDesc("");
    } catch (error) {
      // Handled in context
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button size="sm">
          <Icon name="Plus" className="mr-2 h-4 w-4" />
          New Access Right
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create Access Right</DialogTitle>
          <DialogDescription>
            Grouping endpoints makes it easier to manage permissions.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. User Management"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="desc">Description</Label>
            <Input
              id="desc"
              value={desc}
              onChange={(e) => setDesc(e.target.value)}
              placeholder="Manage all user related operations"
            />
          </div>
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setIsOpen(false)}
            disabled={isCreating}
          >
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isCreating || !name}>
            {isCreating && (
              <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
            )}
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
