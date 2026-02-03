"use client";

import { useState, useEffect, useCallback } from "react";
import { Button } from "~/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { Icon } from "~/components/shared/icon";
import { Input } from "~/components/ui/input";
import { Badge } from "~/components/ui/badge";
import { auditApi, AuditLog } from "~/lib/api/audit";
import { toast } from "sonner";
import { LogDetailDialog } from "~/components/dashboard/audit/log-detail-dialog";
import {
    Pagination,
    PaginationContent,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
  } from "~/components/ui/pagination";

export default function AuditPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [page, setPage] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const pageSize = 15;

  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [isDetailOpen, setIsDetailOpen] = useState(false);

  const fetchLogs = useCallback(async () => {
    setIsLoading(true);
    try {
      const filter: any = {
        page: page,
        page_size: pageSize,
        sort: [{ colId: "created_at", sort: "desc" }]
      };

      if (searchTerm) {
        filter.filter = {
          action: { type: "contains", filter: searchTerm }
        };
      }

      const response = await auditApi.search(filter);
      if (response && response.data) {
        setLogs(response.data);
        setTotalItems(response.paging?.total || 0);
      } else {
        setLogs([]);
        setTotalItems(0);
      }
    } catch (error) {
      console.error("Failed to fetch audit logs:", error);
      toast.error("Failed to fetch audit logs");
    } finally {
      setIsLoading(false);
    }
  }, [page, searchTerm]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const totalPages = Math.ceil(totalItems / pageSize);

  const handleRowClick = (log: AuditLog) => {
    setSelectedLog(log);
    setIsDetailOpen(true);
  };

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
          <Button variant="outline" size="sm" onClick={() => window.open(auditApi.export(), '_blank')}>
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
                setPage(1); // Reset to first page on search
            }}
            className="h-8 w-[150px] lg:w-[250px]"
          />
          {isLoading && <Icon name="Loader" className="h-4 w-4 animate-spin text-muted-foreground" />}
        </div>
        <div className="text-xs text-muted-foreground">
            Total: {totalItems} logs
        </div>
      </div>

      <div className="rounded-md border bg-card">
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
                <TableCell colSpan={6} className="h-24 text-center text-muted-foreground italic">
                  No logs found.
                </TableCell>
              </TableRow>
            ) : (
              logs.map((log) => (
                <TableRow 
                    key={log.id} 
                    className="cursor-pointer hover:bg-muted/50 transition-colors group"
                    onClick={() => handleRowClick(log)}
                >
                  <TableCell className="text-muted-foreground text-xs whitespace-nowrap">
                    {new Date(log.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className="font-mono text-[10px] uppercase bg-primary/5 text-primary border-primary/10">
                        {log.action}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-medium text-xs">{log.entity}</TableCell>
                  <TableCell className="font-mono text-[10px] text-muted-foreground truncate max-w-[120px]">
                    {log.entity_id}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">{log.ip_address}</TableCell>
                  <TableCell>
                    <Icon name="ChevronRight" className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {totalPages > 1 && (
        <div className="mt-4">
          <Pagination>
            <PaginationContent>
              <PaginationItem>
                <Button 
                    variant="ghost" 
                    size="sm" 
                    disabled={page === 1}
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    className="gap-1 pl-2.5"
                >
                  <Icon name="ChevronLeft" className="h-4 w-4" />
                  <span>Previous</span>
                </Button>
              </PaginationItem>
              
              {Array.from({ length: Math.min(5, totalPages) }).map((_, i) => {
                // Simple pagination logic for first 5 pages
                const pageNum = i + 1;
                return (
                  <PaginationItem key={pageNum}>
                    <PaginationLink 
                        isActive={page === pageNum}
                        onClick={() => setPage(pageNum)}
                        className="cursor-pointer"
                    >
                      {pageNum}
                    </PaginationLink>
                  </PaginationItem>
                );
              })}
              
              {totalPages > 5 && <PaginationItem><span className="text-muted-foreground">...</span></PaginationItem>}

              <PaginationItem>
                <Button 
                    variant="ghost" 
                    size="sm" 
                    disabled={page === totalPages}
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                    className="gap-1 pr-2.5"
                >
                  <span>Next</span>
                  <Icon name="ChevronRight" className="h-4 w-4" />
                </Button>
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </div>
      )}

      <LogDetailDialog 
        log={selectedLog}
        open={isDetailOpen}
        onOpenChange={setIsDetailOpen}
      />
    </div>
  );
}