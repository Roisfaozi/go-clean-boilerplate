"use client";

import { useAudit } from "./audit-context";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { Input } from "~/components/ui/input";
import { auditApi } from "~/lib/api/audit";

export function AuditToolbar() {
  const { searchTerm, setSearchTerm, setPage, isLoading, totalItems } = useAudit();

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Audit Logs</h2>
          <p className="text-muted-foreground">
            Monitor system activity and user actions.
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => window.open(auditApi.export(), "_blank")}
          >
            <Icon name="Download" className="mr-2 h-4 w-4" />
            Export CSV
          </Button>
        </div>
      </div>

      <div className="flex items-center justify-between">
        <div className="flex flex-1 items-center space-x-2">
          <Input
            placeholder="Search by action..."
            value={searchTerm}
            onChange={(e) => {
              setSearchTerm(e.target.value);
              setPage(1);
            }}
            className="h-8 w-[150px] lg:w-[250px]"
          />
          {isLoading && (
            <Icon
              name="Loader"
              className="text-muted-foreground h-4 w-4 animate-spin"
            />
          )}
        </div>
        <div className="text-muted-foreground text-xs">
          Total: {totalItems} logs
        </div>
      </div>
    </div>
  );
}
