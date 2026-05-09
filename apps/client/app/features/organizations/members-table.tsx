import { useState } from "react";
import {  NexusCard  } from "@casbin/ui";
import {  NexusInput  } from "@casbin/ui";
import {  NexusButton  } from "@casbin/ui";
import {  Avatar, AvatarFallback, AvatarImage  } from "@casbin/ui";
import {  Badge  } from "@casbin/ui";
import { MemberRoleSelector, MemberRoleBadge } from "./member-role-selector";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@casbin/ui";
import {  useToast  } from "@casbin/ui";
import { Search, MoreHorizontal, UserMinus, Mail } from "lucide-react";
import { cn } from "@casbin/ui";

interface Member {
  id: string;
  name: string;
  email: string;
  avatar?: string;
  role: string;
  status: "online" | "away" | "offline";
  joinedAt: string;
}

const mockMembers: Member[] = [
  { id: "1", name: "Alice Johnson", email: "alice@acme.com", role: "owner", status: "online", joinedAt: "2024-01-15" },
  { id: "2", name: "Bob Smith", email: "bob@acme.com", role: "admin", status: "online", joinedAt: "2024-02-20" },
  { id: "3", name: "Carol Williams", email: "carol@acme.com", role: "editor", status: "away", joinedAt: "2024-03-01" },
  { id: "4", name: "David Brown", email: "david@acme.com", role: "editor", status: "offline", joinedAt: "2024-03-05" },
  { id: "5", name: "Eva Martinez", email: "eva@acme.com", role: "viewer", status: "online", joinedAt: "2024-03-08" },
  { id: "6", name: "Frank Lee", email: "frank@acme.com", role: "viewer", status: "offline", joinedAt: "2024-03-10" },
];

const statusColors: Record<string, string> = {
  online: "bg-success",
  away: "bg-warning",
  offline: "bg-muted-foreground/40",
};

export function MembersTable() {
  const { toast } = useToast();
  const [search, setSearch] = useState("");
  const [members, setMembers] = useState<Member[]>(mockMembers);

  const filtered = members.filter(
    (m) =>
      m.name.toLowerCase().includes(search.toLowerCase()) ||
      m.email.toLowerCase().includes(search.toLowerCase())
  );

  const handleRoleChange = (memberId: string, newRole: string) => {
    setMembers((prev) =>
      prev.map((m) => (m.id === memberId ? { ...m, role: newRole } : m))
    );
    toast({ title: "Role Updated", description: `Changed to ${newRole}` });
  };

  const handleRemove = (member: Member) => {
    setMembers((prev) => prev.filter((m) => m.id !== member.id));
    toast({ title: "Member Removed", description: member.email });
  };

  const addMember = (email: string, role: string) => {
    const newMember: Member = {
      id: String(Date.now()),
      name: email.split("@")[0],
      email,
      role,
      status: "offline",
      joinedAt: new Date().toISOString().slice(0, 10),
    };
    setMembers((prev) => [...prev, newMember]);
  };

  return { filtered, search, setSearch, handleRoleChange, handleRemove, addMember, members };
}

export function MembersTableUI() {
  const { filtered, search, setSearch, handleRoleChange, handleRemove } = MembersTable();

  return (
    <NexusCard className="overflow-hidden">
      <div className="p-4 border-b border-border">
        <div className="relative max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <NexusInput
            placeholder="Search members..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9 h-9"
          />
        </div>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border bg-muted/50">
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">Member</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">Status</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">Role</th>
              <th className="text-left px-4 py-3 font-medium text-muted-foreground">Joined</th>
              <th className="text-right px-4 py-3 font-medium text-muted-foreground w-12"></th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((member) => (
              <tr key={member.id} className="border-b border-border last:border-0 hover:bg-muted/20 transition-colors">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <div className="relative">
                      <Avatar className="h-8 w-8">
                        <AvatarImage src={member.avatar} />
                        <AvatarFallback className="text-xs bg-primary/10 text-primary">
                          {member.name.split(" ").map((n) => n[0]).join("")}
                        </AvatarFallback>
                      </Avatar>
                      <span
                        className={cn(
                          "absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-background",
                          statusColors[member.status]
                        )}
                      />
                    </div>
                    <div>
                      <p className="font-medium text-foreground">{member.name}</p>
                      <p className="text-xs text-muted-foreground">{member.email}</p>
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3">
                  <Badge
                    variant={member.status === "online" ? "default" : "outline"}
                    className={cn(
                      "text-[10px] capitalize",
                      member.status === "online" && "bg-success/10 text-success border-success/20",
                      member.status === "away" && "bg-warning/10 text-warning border-warning/20"
                    )}
                  >
                    {member.status}
                  </Badge>
                </td>
                <td className="px-4 py-3">
                  <MemberRoleSelector
                    value={member.role}
                    onChange={(val) => handleRoleChange(member.id, val)}
                    disabled={member.role === "owner"}
                  />
                </td>
                <td className="px-4 py-3 text-muted-foreground text-xs">{member.joinedAt}</td>
                <td className="px-4 py-3 text-right">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <NexusButton variant="ghost" size="icon" className="h-8 w-8">
                        <MoreHorizontal className="h-4 w-4" />
                      </NexusButton>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem>
                        <Mail className="h-4 w-4 mr-2" />
                        Resend Invite
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="text-destructive focus:text-destructive"
                        onClick={() => handleRemove(member)}
                        disabled={member.role === "owner"}
                      >
                        <UserMinus className="h-4 w-4 mr-2" />
                        Remove Member
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </td>
              </tr>
            ))}
            {filtered.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">
                  No members found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </NexusCard>
  );
}
