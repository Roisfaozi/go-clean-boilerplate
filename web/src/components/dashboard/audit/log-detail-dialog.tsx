"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { AuditLog } from "~/lib/api/audit";
import { ScrollArea } from "~/components/ui/scroll-area";
import { Badge } from "~/components/ui/badge";
import { Icon } from "~/components/shared/icon";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";

interface LogDetailDialogProps {
  log: AuditLog | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function LogDetailDialog({
  log,
  open,
  onOpenChange,
}: LogDetailDialogProps) {
  if (!log) return null;

  const formatDate = (timestamp: number) => {
    return new Date(timestamp).toLocaleString();
  };

  const JsonViewer = ({ data, title }: { data: any; title: string }) => {
    const jsonStr =
      typeof data === "string" ? data : JSON.stringify(data, null, 2);
    const isEmpty =
      !data || jsonStr === "{}" || jsonStr === "[]" || jsonStr === "null";

    return (
      <div className="space-y-2">
        <h4 className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          {title}
        </h4>
        <div className="bg-muted max-h-[300px] overflow-auto rounded-md p-4 font-mono text-xs">
          {isEmpty ? (
            <span className="text-muted-foreground italic">
              No data available
            </span>
          ) : (
            <pre className="break-all whitespace-pre-wrap">{jsonStr}</pre>
          )}
        </div>
      </div>
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[90vh] flex-col sm:max-w-[600px]">
        <DialogHeader>
          <div className="mb-1 flex items-center gap-2">
            <Badge
              variant="outline"
              className="bg-primary/5 text-primary border-primary/10 uppercase"
            >
              {log.action}
            </Badge>
            <span className="text-muted-foreground text-xs">
              {formatDate(log.created_at)}
            </span>
          </div>
          <DialogTitle className="flex items-center gap-2 text-xl">
            <Icon name="FileText" className="text-muted-foreground h-5 w-5" />
            Log Details
          </DialogTitle>
          <DialogDescription>
            Detailed information for audit log ID:{" "}
            <span className="font-mono text-[10px]">{log.id}</span>
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="mt-4 flex-1 pr-4">
          <div className="space-y-6">
            {/* Summary Grid */}
            <div className="bg-muted/30 grid grid-cols-2 gap-4 rounded-lg border p-4">
              <div className="space-y-1">
                <span className="text-muted-foreground text-[10px] font-bold uppercase">
                  Entity
                </span>
                <p className="text-sm font-medium">{log.entity}</p>
              </div>
              <div className="space-y-1">
                <span className="text-muted-foreground text-[10px] font-bold uppercase">
                  Entity ID
                </span>
                <p className="truncate font-mono text-sm" title={log.entity_id}>
                  {log.entity_id}
                </p>
              </div>
              <div className="space-y-1">
                <span className="text-muted-foreground text-[10px] font-bold uppercase">
                  User ID
                </span>
                <p className="truncate font-mono text-sm" title={log.user_id}>
                  {log.user_id}
                </p>
              </div>
              <div className="space-y-1">
                <span className="text-muted-foreground text-[10px] font-bold uppercase">
                  IP Address
                </span>
                <p className="text-sm">{log.ip_address}</p>
              </div>
            </div>

            <div className="space-y-1 px-1">
              <span className="text-muted-foreground text-[10px] font-bold uppercase">
                User Agent
              </span>
              <p className="text-muted-foreground bg-muted/20 rounded p-2 text-xs break-all">
                {log.user_agent}
              </p>
            </div>

            <Tabs defaultValue="changes" className="w-full">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="changes">Changes</TabsTrigger>
                <TabsTrigger value="raw">Raw Data</TabsTrigger>
              </TabsList>
              <TabsContent value="changes" className="space-y-4 pt-4">
                <JsonViewer title="Old Values" data={log.old_values} />
                <JsonViewer title="New Values" data={log.new_values} />
              </TabsContent>
              <TabsContent value="raw" className="pt-4">
                <JsonViewer title="Full Object" data={log} />
              </TabsContent>
            </Tabs>
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}
