"use client";

import { useDashboard } from "./dashboard-context";
import { Badge } from "~/components/ui/badge";
import { ActivityChart } from "~/components/dashboard/activity-chart";

export function SystemInsights() {
  const { insights, isLoading } = useDashboard();

  return (
    <div className="grid grid-cols-1 gap-[var(--spacing-gap)] lg:grid-cols-2">
      <ActivityChart />
      <div className="bg-card text-card-foreground rounded-[var(--radius-lg)] border p-6 shadow-sm">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-primary text-lg font-semibold tracking-tight">
            System Insights
          </h2>
          <Badge variant="outline" className="bg-primary/5">
            Experimental
          </Badge>
        </div>
        <div className="space-y-4">
          <div className="bg-muted/30 rounded-lg border border-dashed p-4">
            <p className="text-muted-foreground text-sm leading-relaxed italic">
              {insights
                ? `User engagement is stable. Most active role currently: ${insights.most_active_role}.`
                : "Analyzing system performance and user engagement patterns..."}
            </p>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="rounded-md border p-3">
              <span className="text-muted-foreground text-[10px] font-bold uppercase">
                Latency
              </span>
              <div className="font-mono text-xl">
                {isLoading ? "..." : `${insights?.avg_latency_ms || 0}ms`}
              </div>
            </div>
            <div className="rounded-md border p-3">
              <span className="text-muted-foreground text-[10px] font-bold uppercase">
                Errors
              </span>
              <div className="font-mono text-xl text-emerald-500">
                {isLoading
                  ? "..."
                  : `${((insights?.error_rate || 0) * 100).toFixed(1)}%`}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
