"use client";

import { useAudit } from "./audit-context";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Badge } from "~/components/ui/badge";
import { Icon } from "~/components/shared/icon";
import { LogDetailDialog } from "~/components/dashboard/audit/log-detail-dialog";
import { AuditLog } from "~/lib/api/audit";
import { memo } from "react";

export function AuditTable() {
  const { logs, isLoading, handleRowClick, selectedLog, isDetailOpen, setIsDetailOpen } = useAudit();

  return (
    <>
      <div className="bg-card rounded-md border">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/50">
              <TableHead className="w-[180px]">Timestamp</TableHead>
              <TableHead>Action</TableHead>
              <TableHead>Entity</TableHead>
              <TableHead>Entity ID</TableHead>
              <TableHead>IP Address</TableHead>
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading && logs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="h-24 text-center">
                  <div className="flex items-center justify-center gap-2">
                    <Icon name="Loader" className="h-4 w-4 animate-spin" />
                    Loading logs...
                  </div>
                </TableCell>
              </TableRow>
            ) : logs.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  className="text-muted-foreground h-24 text-center italic"
                >
                  No logs found.
                </TableCell>
              </TableRow>
            ) : (
              logs.map((log) => (
                <MemoizedAuditRow
                  key={log.id}
                  log={log}
                  onClick={() => handleRowClick(log)}
                />
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <LogDetailDialog
        log={selectedLog}
        open={isDetailOpen}
        onOpenChange={setIsDetailOpen}
      />
    </>
  );
}

const MemoizedAuditRow = memo(function AuditRow({
  log,
  onClick,
}: {
  log: AuditLog;
  onClick: () => void;
}) {
  return (
    <TableRow
      className="hover:bg-muted/50 group cursor-pointer transition-colors"
      onClick={onClick}
    >
      <TableCell className="text-muted-foreground text-xs whitespace-nowrap">
        {new Date(log.created_at).toLocaleString()}
      </TableCell>
      <TableCell>
        <Badge
          variant="outline"
          className="bg-primary/5 text-primary border-primary/10 font-mono text-[10px] uppercase"
        >
          {log.action}
        </Badge>
      </TableCell>
      <TableCell className="text-xs font-medium">{log.entity}</TableCell>
      <TableCell className="text-muted-foreground max-w-[120px] truncate font-mono text-[10px]">
        {log.entity_id}
      </TableCell>
      <TableCell className="text-muted-foreground text-xs">
        {log.ip_address}
      </TableCell>
      <TableCell>
        <Icon
          name="ChevronRight"
          className="text-muted-foreground h-4 w-4 opacity-0 transition-opacity group-hover:opacity-100"
        />
      </TableCell>
    </TableRow>
  );
});
