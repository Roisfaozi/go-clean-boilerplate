"use client";

import { useOrganizationStore } from "~/stores/use-organization-store";
import { Icon } from "~/components/shared/icon";
import { MembersProvider } from "./_components/members-context";
import { MemberInviteDialog } from "./_components/member-invite-dialog";
import { MemberTable } from "./_components/member-list";

export default function OrganizationMembersPage() {
  const { currentOrganization } = useOrganizationStore();

  if (!currentOrganization) {
    return (
      <div className="flex h-[400px] items-center justify-center rounded-lg border-2 border-dashed">
        <div className="text-center">
          <Icon
            name="Building2"
            className="text-muted-foreground/50 mx-auto h-8 w-8"
          />
          <p className="text-muted-foreground mt-2">
            Please select an organization first.
          </p>
        </div>
      </div>
    );
  }

  return (
    <MembersProvider>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">
              Organization Members
            </h2>
            <p className="text-muted-foreground">
              Manage who has access to{" "}
              <strong>{currentOrganization.name}</strong>.
            </p>
          </div>
          <MemberInviteDialog />
        </div>

        <MemberTable />
      </div>
    </MembersProvider>
  );
}
