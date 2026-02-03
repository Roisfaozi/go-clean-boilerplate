"use client";

import { useEffect, useState } from "react";
import { KPICard } from "~/components/dashboard/kpi-card";
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
import { usersApi } from "~/lib/api/users";
import { rolesApi } from "~/lib/api/roles";
import { auditApi, AuditLog } from "~/lib/api/audit";
import Link from "next/link";
import { toast } from "sonner";
import { Badge } from "~/components/ui/badge";

export default function DashboardPage() {
  const [stats, setStats] = useState({
    users: 0,
    roles: 0,
    auditLogs: 0,
  });
  const [recentLogs, setRecentLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        // Parallel data fetching
        const [usersResp, rolesResp, auditResp, recentLogsResp] = await Promise.all([
          usersApi.search({ page: 1, page_size: 1 }), // Just need count
          rolesApi.search({ page: 1, page_size: 1 }), // Just need count
          auditApi.search({ page: 1, page_size: 1 }), // Just need count
          auditApi.search({ 
             page: 1, 
             page_size: 5, 
             sort: [{ colId: "created_at", sort: "desc" }] 
          }),
        ]);

        setStats({
          users: usersResp.paging?.total || 0,
          roles: rolesResp.paging?.total || 0,
          auditLogs: auditResp.paging?.total || 0,
        });

        if (recentLogsResp.data) {
          setRecentLogs(recentLogsResp.data);
        }

      } catch (error) {
        console.error("Dashboard fetch error:", error);
        toast.error("Failed to load dashboard data");
      } finally {
        setIsLoading(false);
      }
    };

    fetchDashboardData();
  }, []);

  const formatTimeAgo = (timestamp: number) => {
    const diff = Date.now() - timestamp;
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d ago`;
    if (hours > 0) return `${hours}h ago`;
    if (minutes > 0) return `${minutes}m ago`;
    return "Just now";
  };

  return (
    <div className="space-y-[var(--spacing-gap)]">
      {/* Zone A: KPI Cards */}
      <div className="grid gap-[var(--spacing-gap)] grid-cols-1 md:grid-cols-2 lg:grid-cols-4">
        <KPICard
          title="Total Users"
          value={isLoading ? "..." : stats.users.toLocaleString()}
          trend={isLoading ? "" : "Active users"}
          trendUp={true}
          iconName="Users"
          description="Registered accounts"
        />
        <KPICard
          title="Defined Roles"
          value={isLoading ? "..." : stats.roles.toLocaleString()}
          trend={isLoading ? "" : "RBAC Policies"}
          trendUp={true}
          iconName="Shield"
          description="Access control roles"
        />
        <KPICard
          title="Total Events"
          value={isLoading ? "..." : stats.auditLogs.toLocaleString()}
          trend={isLoading ? "" : "System logs"}
          trendUp={true}
          iconName="FileText"
          description="Recorded audit trails"
        />
        <KPICard
          title="System Status"
          value="Healthy"
          trend="All systems go"
          trendUp={true}
          iconName="Activity"
          description="No incidents reported"
        />
      </div>

      {/* Zone B & C: Main Content + Quick Actions */}
      <div className="grid gap-[var(--spacing-gap)] md:grid-cols-7">
        
        {/* Recent Activity Table (Span 5) */}
        <div className="md:col-span-5 flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold tracking-tight">Recent Activity</h2>
            <Link href="/dashboard/audit">
              <Button variant="ghost" size="sm" className="gap-1">
                View All <Icon name="ArrowRight" className="h-4 w-4" />
              </Button>
            </Link>
          </div>
          
          <div className="rounded-[var(--radius-lg)] border bg-card text-card-foreground shadow-sm overflow-hidden">
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
                   Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      <TableCell><div className="h-4 w-24 bg-muted/50 rounded animate-pulse" /></TableCell>
                      <TableCell><div className="h-4 w-16 bg-muted/50 rounded animate-pulse" /></TableCell>
                      <TableCell><div className="h-4 w-20 bg-muted/50 rounded animate-pulse" /></TableCell>
                      <TableCell><div className="h-4 w-24 bg-muted/50 rounded animate-pulse" /></TableCell>
                      <TableCell className="text-right"><div className="h-4 w-12 bg-muted/50 rounded animate-pulse ml-auto" /></TableCell>
                    </TableRow>
                   ))
                ) : recentLogs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                      No recent activity found.
                    </TableCell>
                  </TableRow>
                ) : (
                  recentLogs.map((log) => (
                    <TableRow key={log.id}>
                      <TableCell className="font-medium text-xs">{log.user_id}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className="text-[10px] font-mono uppercase bg-muted/50">
                          {log.action}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-muted-foreground text-xs">{log.entity}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{log.ip_address}</TableCell>
                      <TableCell className="text-right text-muted-foreground text-xs">{formatTimeAgo(log.created_at)}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>

        {/* Quick Actions (Span 2) */}
        <div className="md:col-span-2 flex flex-col gap-4">
          <h2 className="text-lg font-semibold tracking-tight">Quick Actions</h2>
          <div className="grid gap-3">
            <Link href="/dashboard/users">
                <Button className="w-full justify-start h-auto py-4" variant="outline">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10 rounded-md text-primary">
                    <Icon name="UserPlus" className="h-5 w-5" />
                    </div>
                    <div className="text-left">
                    <div className="font-semibold">Manage Users</div>
                    <div className="text-xs text-muted-foreground">Add or edit accounts</div>
                    </div>
                </div>
                </Button>
            </Link>
            
            <Link href="/dashboard/roles">
                <Button className="w-full justify-start h-auto py-4" variant="outline">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-accent/10 rounded-md text-accent">
                    <Icon name="Shield" className="h-5 w-5" />
                    </div>
                    <div className="text-left">
                    <div className="font-semibold">Configure Roles</div>
                    <div className="text-xs text-muted-foreground">Update permissions</div>
                    </div>
                </div>
                </Button>
            </Link>

            <Button 
                className="w-full justify-start h-auto py-4" 
                variant="outline"
                onClick={() => window.open(auditApi.export(), '_blank')}
            >
              <div className="flex items-center gap-3">
                <div className="p-2 bg-secondary/10 rounded-md text-secondary">
                  <Icon name="Download" className="h-5 w-5" />
                </div>
                <div className="text-left">
                  <div className="font-semibold">Export Logs</div>
                  <div className="text-xs text-muted-foreground">Download audit trail</div>
                </div>
              </div>
            </Button>
          </div>
        </div>

      </div>
    </div>
  );
}