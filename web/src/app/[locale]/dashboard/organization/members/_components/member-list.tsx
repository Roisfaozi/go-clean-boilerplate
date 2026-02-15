"use client";

import { ReactNode } from "react";
import { useMembers } from "./members-context";
import { useOrganizationStore } from "~/stores/use-organization-store";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import { Badge } from "~/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { Member } from "~/lib/api/organizations";
import { memo } from "react";

export function MemberTable() {
  const { members, isLoading } = useMembers();

  return (
    <div className="bg-card rounded-md border">
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
            <MemberTableSkeleton />
          ) : members.length === 0 ? (
            <MemberTableEmpty />
          ) : (
            members.map((member) => (
              <MemoizedMemberTableRow key={member.id} member={member} />
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

const MemoizedMemberTableRow = memo(function MemberTableRow({
  member,
}: {
  member: Member;
}) {
  const { currentOrganization } = useOrganizationStore();
  const { roles, updateMemberRole, removeMember } = useMembers();

  if (!currentOrganization) return null;

  return (
    <TableRow>
      <TableCell>
        <div className="flex items-center gap-3">
          <Avatar className="h-8 w-8 border">
            <AvatarImage src={member.user?.avatar_url} />
            <AvatarFallback>
              {member.user?.name?.[0] || member.user?.email?.[0].toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <div className="flex flex-col">
            <span className="text-sm font-medium">
              {member.user?.name || "Invited User"}
            </span>
            <span className="text-muted-foreground text-[10px]">
              {member.user?.email}
            </span>
          </div>
        </div>
      </TableCell>
      <TableCell>
        <Select
          defaultValue={member.role_id}
          onValueChange={(val) => updateMemberRole(member.user_id, val)}
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
        <Badge
          variant={member.status === "active" ? "success" : "secondary"}
          className="text-[10px] uppercase"
        >
          {member.status}
        </Badge>
      </TableCell>
      <TableCell className="text-muted-foreground text-xs">
        {member.joined_at
          ? new Date(member.joined_at).toLocaleDateString()
          : "-"}
      </TableCell>
      <TableCell className="text-right">
        <Button
          variant="ghost"
          size="icon"
          className="text-destructive hover:bg-destructive/10 h-8 w-8"
          onClick={() =>
            removeMember(
              member.user_id,
              member.user?.name || member.user?.email || ""
            )
          }
          disabled={member.user_id === currentOrganization.owner_id}
        >
          <Icon name="UserMinus" className="h-4 w-4" />
        </Button>
      </TableCell>
    </TableRow>
  );
});

function MemberTableSkeleton() {
  return (
    <>
      {Array.from({ length: 3 }).map((_, i) => (
        <TableRow key={i}>
          <TableCell colSpan={5} className="bg-muted/10 h-16 animate-pulse" />
        </TableRow>
      ))}
    </>
  );
}

function MemberTableEmpty() {
  return (
    <TableRow>
      <TableCell
        colSpan={5}
        className="text-muted-foreground h-24 text-center italic"
      >
        No members found.
      </TableCell>
    </TableRow>
  );
}
