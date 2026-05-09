import { useState } from "react";
import { Badge } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@casbin/ui";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@casbin/ui";
import { Search, Download, Filter, ChevronLeft, ChevronRight } from "lucide-react";

export interface AuditLogEntry {
  id: string;
  action: string;
  actor: string;
  target: string;
  ip_address: string;
  timestamp: string;
  severity: "info" | "warning" | "critical";
  details?: string;
}

const severityMap: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  info: "secondary",
  warning: "outline",
  critical: "destructive",
};

const severityDot: Record<string, string> = {
  info: "bg-info",
  warning: "bg-warning",
  critical: "bg-destructive",
};

interface AuditLogTableProps {
  logs: AuditLogEntry[];
}

export function AuditLogTable({ logs }: AuditLogTableProps) {
  const [search, setSearch] = useState("");
  const [severityFilter, setSeverityFilter] = useState<string>("all");
  const [page, setPage] = useState(1);
  const perPage = 10;

  const filtered = logs.filter((log) => {
    const matchesSearch =
      !search ||
      log.action.toLowerCase().includes(search.toLowerCase()) ||
      log.actor.toLowerCase().includes(search.toLowerCase()) ||
      log.target.toLowerCase().includes(search.toLowerCase());
    const matchesSeverity = severityFilter === "all" || log.severity === severityFilter;
    return matchesSearch && matchesSeverity;
  });

  const totalPages = Math.ceil(filtered.length / perPage);
  const paginated = filtered.slice((page - 1) * perPage, page * perPage);

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
        <div className="flex flex-1 items-center gap-2">
          <div className="relative max-w-sm flex-1">
            <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
            <NexusInput
              placeholder="Search logs..."
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              className="pl-9"
            />
          </div>
          <Select
            value={severityFilter}
            onValueChange={(v) => {
              setSeverityFilter(v);
              setPage(1);
            }}
          >
            <SelectTrigger className="w-[140px]">
              <Filter className="text-muted-foreground mr-2 h-3.5 w-3.5" />
              <SelectValue placeholder="Severity" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="info">Info</SelectItem>
              <SelectItem value="warning">Warning</SelectItem>
              <SelectItem value="critical">Critical</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <NexusButton variant="outline" size="sm">
          <Download className="mr-2 h-4 w-4" />
          Export
        </NexusButton>
      </div>

      {/* Table */}
      <div className="border-border overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/50">
              <TableHead>Action</TableHead>
              <TableHead>Actor</TableHead>
              <TableHead>Target</TableHead>
              <TableHead>Severity</TableHead>
              <TableHead>IP Address</TableHead>
              <TableHead>Timestamp</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginated.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-muted-foreground py-8 text-center">
                  No logs found
                </TableCell>
              </TableRow>
            ) : (
              paginated.map((log) => (
                <TableRow key={log.id}>
                  <TableCell className="font-mono text-xs">{log.action}</TableCell>
                  <TableCell>{log.actor}</TableCell>
                  <TableCell className="text-muted-foreground">{log.target}</TableCell>
                  <TableCell>
                    <Badge variant={severityMap[log.severity] ?? "default"} className="gap-1.5">
                      <span className={`h-1.5 w-1.5 rounded-full ${severityDot[log.severity]}`} />
                      {log.severity}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground font-mono text-xs">
                    {log.ip_address}
                  </TableCell>
                  <TableCell className="text-muted-foreground whitespace-nowrap">
                    {log.timestamp}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="text-muted-foreground flex items-center justify-between text-sm">
          <span>
            Showing {(page - 1) * perPage + 1}–{Math.min(page * perPage, filtered.length)} of{" "}
            {filtered.length}
          </span>
          <div className="flex items-center gap-1">
            <NexusButton
              variant="ghost"
              size="sm"
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
            >
              <ChevronLeft className="h-4 w-4" />
            </NexusButton>
            {Array.from({ length: totalPages }, (_, i) => (
              <NexusButton
                key={i}
                variant={page === i + 1 ? "primary" : "ghost"}
                size="sm"
                onClick={() => setPage(i + 1)}
                className="h-8 w-8 p-0"
              >
                {i + 1}
              </NexusButton>
            ))}
            <NexusButton
              variant="ghost"
              size="sm"
              disabled={page >= totalPages}
              onClick={() => setPage(page + 1)}
            >
              <ChevronRight className="h-4 w-4" />
            </NexusButton>
          </div>
        </div>
      )}
    </div>
  );
}
