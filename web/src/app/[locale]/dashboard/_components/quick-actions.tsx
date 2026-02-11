"use client";

import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { auditApi } from "~/lib/api/audit";
import Link from "next/link";

export function QuickActions() {
  return (
    <div className="flex flex-col gap-4 md:col-span-2">
      <h2 className="text-lg font-semibold tracking-tight">Quick Actions</h2>
      <div className="grid gap-3">
        <Link href="/dashboard/users" className="block w-full">
          <Button
            className="h-auto w-full justify-start py-4"
            variant="outline"
          >
            <div className="flex items-center gap-3">
              <div className="bg-primary/10 text-primary rounded-md p-2">
                <Icon name="UserPlus" className="h-5 w-5" />
              </div>
              <div className="text-left">
                <div className="font-semibold">Manage Users</div>
                <div className="text-muted-foreground text-xs">
                  Add or edit accounts
                </div>
              </div>
            </div>
          </Button>
        </Link>

        <Link href="/dashboard/roles" className="block w-full">
          <Button
            className="h-auto w-full justify-start py-4"
            variant="outline"
          >
            <div className="flex items-center gap-3">
              <div className="bg-accent/10 text-accent rounded-md p-2">
                <Icon name="Shield" className="h-5 w-5" />
              </div>
              <div className="text-left">
                <div className="font-semibold">Configure Roles</div>
                <div className="text-muted-foreground text-xs">
                  Update permissions
                </div>
              </div>
            </div>
          </Button>
        </Link>

        <Button
          className="h-auto w-full justify-start py-4"
          variant="outline"
          onClick={() => window.open(auditApi.export(), "_blank")}
        >
          <div className="flex items-center gap-3">
            <div className="bg-secondary/10 text-secondary rounded-md p-2">
              <Icon name="Download" className="h-5 w-5" />
            </div>
            <div className="text-left">
              <div className="font-semibold">Export Logs</div>
              <div className="text-muted-foreground text-xs">
                Download audit trail
              </div>
            </div>
          </div>
        </Button>
      </div>
    </div>
  );
}
