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

export function LogDetailDialog({ log, open, onOpenChange }: LogDetailDialogProps) {
  if (!log) return null;

  const formatDate = (timestamp: number) => {
    return new Date(timestamp).toLocaleString();
  };

  const JsonViewer = ({ data, title }: { data: any, title: string }) => {
    const jsonStr = typeof data === 'string' ? data : JSON.stringify(data, null, 2);
    const isEmpty = !data || jsonStr === "{}" || jsonStr === "[]" || jsonStr === "null";

    return (
      <div className="space-y-2">
        <h4 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{title}</h4>
        <div className="rounded-md bg-muted p-4 font-mono text-xs overflow-auto max-h-[300px]">
          {isEmpty ? (
            <span className="italic text-muted-foreground">No data available</span>
          ) : (
            <pre className="whitespace-pre-wrap break-all">{jsonStr}</pre>
          )}
        </div>
      </div>
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] flex flex-col">
        <DialogHeader>
          <div className="flex items-center gap-2 mb-1">
             <Badge variant="outline" className="bg-primary/5 text-primary border-primary/10 uppercase">
                {log.action}
             </Badge>
             <span className="text-xs text-muted-foreground">{formatDate(log.created_at)}</span>
          </div>
          <DialogTitle className="text-xl flex items-center gap-2">
            <Icon name="FileText" className="h-5 w-5 text-muted-foreground" />
            Log Details
          </DialogTitle>
          <DialogDescription>
            Detailed information for audit log ID: <span className="font-mono text-[10px]">{log.id}</span>
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className="flex-1 mt-4 pr-4">
          <div className="space-y-6">
            {/* Summary Grid */}
            <div className="grid grid-cols-2 gap-4 border rounded-lg p-4 bg-muted/30">
              <div className="space-y-1">
                <span className="text-[10px] font-bold uppercase text-muted-foreground">Entity</span>
                <p className="text-sm font-medium">{log.entity}</p>
              </div>
              <div className="space-y-1">
                <span className="text-[10px] font-bold uppercase text-muted-foreground">Entity ID</span>
                <p className="text-sm font-mono truncate" title={log.entity_id}>{log.entity_id}</p>
              </div>
              <div className="space-y-1">
                <span className="text-[10px] font-bold uppercase text-muted-foreground">User ID</span>
                <p className="text-sm font-mono truncate" title={log.user_id}>{log.user_id}</p>
              </div>
              <div className="space-y-1">
                <span className="text-[10px] font-bold uppercase text-muted-foreground">IP Address</span>
                <p className="text-sm">{log.ip_address}</p>
              </div>
            </div>

            <div className="space-y-1 px-1">
               <span className="text-[10px] font-bold uppercase text-muted-foreground">User Agent</span>
               <p className="text-xs text-muted-foreground break-all bg-muted/20 p-2 rounded">{log.user_agent}</p>
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
