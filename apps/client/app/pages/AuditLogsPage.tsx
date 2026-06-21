import { useQuery } from "@tanstack/react-query";
import { PageHeader } from "@/components/layout/page-header";
import { AuditLogTable, AuditLogEntry } from "@/components/admin/audit-log-table";
import { auditApi } from "@/lib/api/audit";
import { format } from "date-fns";

export default function AuditLogsPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["audit-logs"],
    queryFn: async () => {
      return await auditApi.search({
        page: 1,
        limit: 100,
        sort: "created_at desc",
      });
    },
  });

  const logs: AuditLogEntry[] =
    data?.data?.map((log) => ({
      id: log.id,
      action: log.action,
      actor: log.user_id, // We might want to fetch usernames later
      target: `${log.entity}:${log.entity_id}`,
      ip_address: log.ip_address,
      timestamp: format(new Date(log.created_at), "yyyy-MM-dd HH:mm:ss"),
      severity:
        log.action.includes("delete") || log.action.includes("failed")
          ? "critical"
          : log.action.includes("update") || log.action.includes("permission")
            ? "warning"
            : "info",
    })) || [];

  return (
    <div className="space-y-6">
      <PageHeader title="Audit Logs" description="System activity and change history" />
      <AuditLogTable logs={logs} />
    </div>
  );
}
