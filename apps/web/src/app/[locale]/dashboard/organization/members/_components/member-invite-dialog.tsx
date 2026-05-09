"use client";

import { useState, ReactNode } from "react";
import { useMembers } from "./members-context";
import { useOrganizationStore } from "~/stores/use-organization-store";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";

export function MemberInviteDialog() {
  const { currentOrganization } = useOrganizationStore();
  const { roles, inviteMember, isInviting } = useMembers();
  const [email, setEmail] = useState("");
  const [roleId, setRoleId] = useState("");
  const [isOpen, setIsOpen] = useState(false);

  const handleSubmit = async () => {
    if (!email || !roleId) return;
    try {
      await inviteMember(email, roleId);
      setIsOpen(false);
      setEmail("");
      setRoleId("");
    } catch (error) {
      // Error handled in context
    }
  };

  if (!currentOrganization) return null;

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button size="sm">
          <Icon name="UserPlus" className="mr-2 h-4 w-4" />
          Invite Member
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite new member</DialogTitle>
          <DialogDescription>
            Invite someone to join {currentOrganization.name} by their email
            address.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="email">Email address</Label>
            <Input
              id="email"
              type="email"
              placeholder="colleague@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="role">Initial Role</Label>
            <Select value={roleId} onValueChange={setRoleId}>
              <SelectTrigger id="role">
                <SelectValue placeholder="Select a role" />
              </SelectTrigger>
              <SelectContent>
                {roles.map((role) => (
                  <SelectItem key={role.id} value={role.id}>
                    {role.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setIsOpen(false)}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={isInviting || !email || !roleId}
          >
            {isInviting && (
              <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
            )}
            Send Invitation
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
