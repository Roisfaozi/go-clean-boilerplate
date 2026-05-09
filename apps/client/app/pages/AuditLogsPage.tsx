import { PageHeader } from "@/components/layout/page-header";
import { AuditLogTable, AuditLogEntry } from "@/components/admin/audit-log-table";

const mockLogs: AuditLogEntry[] = [
  { id: "1", action: "user.created", actor: "admin@nexus.io", target: "john@example.com", ip_address: "192.168.1.10", timestamp: "2024-03-08 14:23:01", severity: "info" },
  { id: "2", action: "role.updated", actor: "admin@nexus.io", target: "Editor", ip_address: "192.168.1.10", timestamp: "2024-03-08 13:45:22", severity: "warning" },
  { id: "3", action: "permission.deleted", actor: "admin@nexus.io", target: "write:reports", ip_address: "192.168.1.10", timestamp: "2024-03-08 12:10:45", severity: "critical" },
  { id: "4", action: "user.login", actor: "jane@example.com", target: "-", ip_address: "10.0.0.42", timestamp: "2024-03-08 11:30:00", severity: "info" },
  { id: "5", action: "org.settings_changed", actor: "admin@nexus.io", target: "Acme Corp", ip_address: "192.168.1.10", timestamp: "2024-03-08 10:05:33", severity: "warning" },
  { id: "6", action: "user.disabled", actor: "admin@nexus.io", target: "bob@example.com", ip_address: "192.168.1.10", timestamp: "2024-03-08 09:45:12", severity: "critical" },
  { id: "7", action: "project.created", actor: "jane@example.com", target: "Project Alpha", ip_address: "10.0.0.42", timestamp: "2024-03-08 09:20:00", severity: "info" },
  { id: "8", action: "role.created", actor: "admin@nexus.io", target: "Auditor", ip_address: "192.168.1.10", timestamp: "2024-03-08 08:55:33", severity: "info" },
  { id: "9", action: "permission.granted", actor: "admin@nexus.io", target: "read:analytics", ip_address: "192.168.1.10", timestamp: "2024-03-08 08:30:10", severity: "warning" },
  { id: "10", action: "user.login_failed", actor: "unknown@test.com", target: "-", ip_address: "203.0.113.5", timestamp: "2024-03-08 08:12:45", severity: "critical" },
  { id: "11", action: "org.member_added", actor: "admin@nexus.io", target: "sarah@example.com", ip_address: "192.168.1.10", timestamp: "2024-03-08 07:50:00", severity: "info" },
  { id: "12", action: "user.password_reset", actor: "jane@example.com", target: "jane@example.com", ip_address: "10.0.0.42", timestamp: "2024-03-07 22:15:30", severity: "warning" },
];

export default function AuditLogsPage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Audit Logs"
        description="System activity and change history"
      />
      <AuditLogTable logs={mockLogs} />
    </div>
  );
}
