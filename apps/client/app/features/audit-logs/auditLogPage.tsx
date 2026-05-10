import { useState, useMemo } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { useAuditLogs, useExportAuditLogs } from "./auditHooks";
import { AuditLogTable, type AuditLogEntry } from "@/components/admin/audit-log-table";
import { NexusButton, NexusCard, NexusInput, Skeleton } from "@casbin/ui";
import { format } from "date-fns";
import { Download, RefreshCw, Search } from "lucide-react";

export default function AuditLogsPage() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const limit = 15;

  const { data: response, isLoading, refetch, isFetching } = useAuditLogs({
    page,
    limit,
    search: search || undefined,
    sort: "created_at desc",
  });

  const exportLogs = useExportAuditLogs();

  const logs: AuditLogEntry[] = useMemo(() => {
    return (
      response?.data?.map((log) => ({
        id: log.id,
        action: log.action,
        actor: log.user_id,
        target: `${log.entity}:${log.entity_id}`,
        ip_address: log.ip_address,
        timestamp: format(new Date(log.created_at), "yyyy-MM-dd HH:mm:ss"),
        severity:
          log.action.includes("delete") || log.action.includes("failed") || log.action.includes("revoke")
            ? "critical"
            : log.action.includes("update") || log.action.includes("permission") || log.action.includes("grant")
              ? "warning"
              : "info",
      })) || []
    );
  }, [response]);

  const totalPages = Math.ceil((response?.meta?.total || 0) / limit);

  return (
    <div className="space-y-6">
      <PageHeader
        title="Audit Logs"
        description="System-wide activity trail and change history."
        actions={
          <div className="flex gap-2">
            <NexusButton
              variant="outline"
              size="sm"
              onClick={() => exportLogs({ format: "csv" })}
            >
              <Download className="h-4 w-4 mr-2" /> Export CSV
            </NexusButton>
            <NexusButton
              variant="outline"
              size="sm"
              onClick={() => refetch()}
              disabled={isFetching}
            >
              <RefreshCw className={`h-4 w-4 mr-2 ${isFetching ? "animate-spin" : ""}`} />
              Refresh
            </NexusButton>
          </div>
        }
      />

      <NexusCard className="p-0 border-none shadow-premium bg-white/50 backdrop-blur-sm overflow-hidden">
        <div className="p-4 border-b bg-muted/30">
          <div className="relative max-w-md">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <NexusInput
              placeholder="Search by action, user, or target..."
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              className="pl-9 bg-white"
            />
          </div>
        </div>

        {isLoading ? (
          <div className="p-6 space-y-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-12 w-full rounded-lg" />
            ))}
          </div>
        ) : (
          <AuditLogTable
            logs={logs}
            // We'll update the table component to handle server-side pagination if needed,
            // but for now it has its own internal pagination which we'll suppress by passing
            // only the current page's data.
          />
        )}

        <div className="p-4 border-t bg-muted/10 flex items-center justify-between">
          <span className="text-sm text-muted-foreground">
            Total Records: {response?.meta?.total || 0}
          </span>
          <div className="flex gap-2">
             <NexusButton
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => setPage(p => p - 1)}
            >
              Previous
            </NexusButton>
            <div className="flex items-center px-4 text-sm font-medium">
              Page {page} of {totalPages || 1}
            </div>
            <NexusButton
              variant="outline"
              size="sm"
              disabled={page >= totalPages}
              onClick={() => setPage(p => p + 1)}
            >
              Next
            </NexusButton>
          </div>
        </div>
      </NexusCard>
    </div>
  );
}
