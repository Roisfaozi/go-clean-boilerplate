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

export function RegisterEpDialog() {
  const { createEndpoint, isCreating } = useAccessRights();
  const [path, setPath] = useState("");
  const [method, setMethod] = useState("GET");
  const [isOpen, setIsOpen] = useState(false);

  const handleSubmit = async () => {
    if (!path) return;
    try {
      await createEndpoint(method, path);
      setIsOpen(false);
      setPath("");
      setMethod("GET");
    } catch (error) {
      // Handled in context
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button size="sm">
          <Icon name="Plus" className="mr-2 h-4 w-4" />
          Register Endpoint
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Register API Endpoint</DialogTitle>
          <DialogDescription>
            Add a new endpoint to the system catalog.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="path">Path</Label>
            <Input
              id="path"
              value={path}
              onChange={(e) => setPath(e.target.value)}
              placeholder="/api/v1/users"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="method">Method</Label>
            <select
              className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-full rounded-md border px-3 py-2 text-sm file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
              value={method}
              onChange={(e) => setMethod(e.target.value)}
            >
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="PATCH">PATCH</option>
              <option value="DELETE">DELETE</option>
            </select>
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
          <Button onClick={handleSubmit} disabled={isCreating || !path}>
            {isCreating && (
              <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
            )}
            Register
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
