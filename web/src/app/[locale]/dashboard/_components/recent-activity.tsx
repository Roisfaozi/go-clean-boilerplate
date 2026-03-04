"use client";

import { useDashboard } from "./dashboard-context";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";

import { Badge } from "~/components/ui/badge";

import { Button } from "~/components/ui/button";

import { Icon } from "~/components/shared/icon";

import Link from "next/link";

import { useEffect, useState, memo } from "react";

import { AuditLog } from "~/lib/api/audit";

import { TableSkeleton } from "~/components/shared/skeletons";

export function RecentActivity() {
  const { recentLogs, isLoading } = useDashboard();

  const [now, setNow] = useState<number | null>(null);

  useEffect(() => {
    const handle = requestAnimationFrame(() => {
      setNow(Date.now());
    });

    return () => cancelAnimationFrame(handle);
  }, []);

  if (isLoading && recentLogs.length === 0) {
    return (
      <div className="space-y-4 md:col-span-5">
        <div className="flex items-center justify-between">
          <div className="bg-muted h-6 w-32 animate-pulse rounded" />
        </div>

        <TableSkeleton rows={5} columns={5} />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 md:col-span-5">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold tracking-tight">
          Recent Activity
        </h2>

        <Link href="/dashboard/audit">
          <Button variant="ghost" size="sm" className="gap-1">
            View All <Icon name="ArrowRight" className="h-4 w-4" />
          </Button>
        </Link>
      </div>

      <div className="bg-card text-card-foreground overflow-hidden rounded-[var(--radius-lg)] border shadow-sm">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>User</TableHead>

              <TableHead>Action</TableHead>

              <TableHead>Entity</TableHead>

              <TableHead>IP</TableHead>

              <TableHead className="text-right">Time</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {isLoading ? (
              <ActivitySkeleton />
            ) : recentLogs.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="text-muted-foreground py-8 text-center"
                >
                  No recent activity found.
                </TableCell>
              </TableRow>
            ) : (
              recentLogs.map((log) => (
                <MemoizedActivityRow key={log.id} log={log} now={now} />
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

const MemoizedActivityRow = memo(function ActivityRow({
  log,
  now,
}: {
  log: AuditLog;
  now: number | null;
}) {
  const formatTimeAgo = (timestamp: number, referenceTime: number | null) => {
    if (!referenceTime) return "...";

    const diff = referenceTime - timestamp;

    const minutes = Math.floor(diff / 60000);

    const hours = Math.floor(minutes / 60);

    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d ago`;

    if (hours > 0) return `${hours}h ago`;

    if (minutes > 0) return `${minutes}m ago`;

    return "Just now";
  };

  return (
    <TableRow>
      <TableCell className="text-xs font-medium">{log.user_id}</TableCell>

      <TableCell>
        <Badge
          variant="outline"
          className="bg-muted/50 font-mono text-[10px] uppercase"
        >
          {log.action}
        </Badge>
      </TableCell>

      <TableCell className="text-muted-foreground text-xs">
        {log.entity}
      </TableCell>

      <TableCell className="text-muted-foreground text-xs">
        {log.ip_address}
      </TableCell>

      <TableCell className="text-muted-foreground text-right text-xs">
        {formatTimeAgo(log.created_at, now)}
      </TableCell>
    </TableRow>
  );
});

function ActivitySkeleton() {
  return (
    <>
      {Array.from({ length: 5 }).map((_, i) => (
        <TableRow key={i}>
          <TableCell>
            <div className="bg-muted/50 h-4 w-24 animate-pulse rounded" />
          </TableCell>
          <TableCell>
            <div className="bg-muted/50 h-4 w-16 animate-pulse rounded" />
          </TableCell>
          <TableCell>
            <div className="bg-muted/50 h-4 w-20 animate-pulse rounded" />
          </TableCell>
          <TableCell>
            <div className="bg-muted/50 h-4 w-24 animate-pulse rounded" />
          </TableCell>
          <TableCell className="text-right">
            <div className="bg-muted/50 ml-auto h-4 w-12 animate-pulse rounded" />
          </TableCell>
        </TableRow>
      ))}
    </>
  );
}
