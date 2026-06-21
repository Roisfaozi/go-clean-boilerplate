import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { Label } from "@casbin/ui";
import { MemberRoleSelector } from "./member-role-selector";
import { useToast } from "@casbin/ui";
import { Plus, Mail } from "lucide-react";

interface InviteMemberModalProps {
  onInvite?: (email: string, role: string) => void;
}

export function InviteMemberModal({ onInvite }: InviteMemberModalProps) {
  const [open, setOpen] = useState(false);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("viewer");
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  const handleInvite = async () => {
    if (!email.trim()) return;
    setLoading(true);
    // Simulate API call
    await new Promise((r) => setTimeout(r, 600));
    onInvite?.(email, role);
    toast({
      title: "Invitation Sent",
      description: `Invited ${email} as ${role}`,
    });
    setEmail("");
    setRole("viewer");
    setLoading(false);
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <NexusButton size="sm">
          <Plus className="mr-1.5 h-4 w-4" />
          Invite Member
        </NexusButton>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Invite Member</DialogTitle>
          <DialogDescription>
            Send an invitation to join the organization
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="invite-email">Email Address</Label>
            <div className="relative">
              <Mail className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
              <NexusInput
                id="invite-email"
                type="email"
                placeholder="colleague@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="pl-9"
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label>Role</Label>
            <MemberRoleSelector value={role} onChange={setRole} />
          </div>
        </div>
        <DialogFooter>
          <NexusButton variant="outline" onClick={() => setOpen(false)}>
            Cancel
          </NexusButton>
          <NexusButton
            onClick={handleInvite}
            disabled={!email.trim() || loading}
          >
            {loading ? "Sending…" : "Send Invitation"}
          </NexusButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
