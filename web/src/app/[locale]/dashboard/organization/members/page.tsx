"use client";

import { useState, useEffect, useCallback } from "react";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { organizationsApi, Member } from "~/lib/api/organizations";
import { rolesApi, Role } from "~/lib/api/roles";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
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
import { Badge } from "~/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import { toast } from "sonner";
import { Skeleton } from "~/components/ui/skeleton";

export default function OrganizationMembersPage() {
  const { currentOrganization } = useOrganizationStore();
  const [members, setMembers] = useState<Member[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Invite state
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRoleId, setInviteRoleId] = useState("");
  const [isInviteOpen, setIsInviteOpen] = useState(false);
  const [isInviting, setIsInviting] = useState(false);

  const fetchData = useCallback(async () => {
    if (!currentOrganization) return;
    setIsLoading(true);
    try {
      const [membersResp, rolesResp] = await Promise.all([
        organizationsApi.getMembers(currentOrganization.id),
        rolesApi.getAll(),
      ]);
      if (membersResp.data) setMembers(membersResp.data);
      if (rolesResp.data) setRoles(rolesResp.data);
    } catch (error) {
      console.error("Failed to fetch data", error);
      toast.error("Failed to load members");
    } finally {
      setIsLoading(false);
    }
  }, [currentOrganization]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleInvite = async () => {
    if (!currentOrganization || !inviteEmail || !inviteRoleId) {
        toast.error("Please fill all fields");
        return;
    }
    setIsInviting(true);
    try {
      await organizationsApi.inviteMember(currentOrganization.id, {
        email: inviteEmail,
        role_id: inviteRoleId,
      });
      toast.success("Invitation sent successfully");
      setIsInviteOpen(false);
      setInviteEmail("");
      setInviteRoleId("");
      fetchData();
    } catch (error: any) {
      toast.error(error.message || "Failed to send invitation");
    } finally {
      setIsInviting(false);
    }
  };

  const handleRemoveMember = async (userId: string, name: string) => {
    if (!currentOrganization) return;
    if (!confirm(`Are you sure you want to remove ${name} from the organization?`)) return;

    try {
      await organizationsApi.removeMember(currentOrganization.id, userId);
      toast.success("Member removed successfully");
      fetchData();
    } catch (error) {
      toast.error("Failed to remove member");
    }
  };

  const handleUpdateRole = async (userId: string, roleId: string) => {
    if (!currentOrganization) return;
    try {
      await organizationsApi.updateMemberRole(currentOrganization.id, userId, { role_id: roleId });
      toast.success("Member role updated");
      fetchData();
    } catch (error) {
      toast.error("Failed to update role");
    }
  };

  if (!currentOrganization) {
    return (
      <div className="flex h-[400px] items-center justify-center rounded-lg border-2 border-dashed">
        <div className="text-center">
          <Icon name="Building2" className="mx-auto h-8 w-8 text-muted-foreground/50" />
          <p className="mt-2 text-muted-foreground">Please select an organization first.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Organization Members</h2>
          <p className="text-muted-foreground">
            Manage who has access to <strong>{currentOrganization.name}</strong>.
          </p>
        </div>

        <Dialog open={isInviteOpen} onOpenChange={setIsInviteOpen}>
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
                Invite someone to join {currentOrganization.name} by their email address.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="email">Email address</Label>
                <Input
                    id="email"
                    type="email"
                    placeholder="colleague@example.com"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="role">Initial Role</Label>
                <Select value={inviteRoleId} onValueChange={setInviteRoleId}>
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
              <Button variant="outline" onClick={() => setIsInviteOpen(false)}>Cancel</Button>
              <Button onClick={handleInvite} disabled={isInviting}>
                {isInviting && <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />}
                Send Invitation
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/50">
              <TableHead>User</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Joined At</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell colSpan={5} className="h-16 animate-pulse bg-muted/10" />
                </TableRow>
              ))
            ) : members.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center text-muted-foreground italic">
                  No members found.
                </TableCell>
              </TableRow>
            ) : (
              members.map((member) => (
                <TableRow key={member.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Avatar className="h-8 w-8 border">
                        <AvatarImage src={member.user?.avatar_url} />
                        <AvatarFallback>{member.user?.name?.[0] || member.user?.email?.[0].toUpperCase()}</AvatarFallback>
                      </Avatar>
                      <div className="flex flex-col">
                        <span className="font-medium text-sm">{member.user?.name || "Invited User"}</span>
                        <span className="text-[10px] text-muted-foreground">{member.user?.email}</span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Select
                        defaultValue={member.role_id}
                        onValueChange={(val) => handleUpdateRole(member.user_id, val)}
                    >
                      <SelectTrigger className="h-8 w-[140px] text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {roles.map((role) => (
                          <SelectItem key={role.id} value={role.id} className="text-xs">
                            {role.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </TableCell>
                  <TableCell>
                    <Badge variant={member.status === "active" ? "success" : "secondary"} className="text-[10px] uppercase">
                      {member.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {member.joined_at ? new Date(member.joined_at).toLocaleDateString() : "-"}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-destructive hover:bg-destructive/10"
                        onClick={() => handleRemoveMember(member.user_id, member.user?.name || member.user?.email || "")}
                        disabled={member.user_id === currentOrganization.owner_id}
                    >
                      <Icon name="UserMinus" className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
