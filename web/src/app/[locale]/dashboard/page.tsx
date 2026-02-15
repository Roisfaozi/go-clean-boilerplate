"use client";

import { DashboardProvider } from "./_components/dashboard-context";
import { DashboardStats } from "./_components/dashboard-stats";
import { SystemInsights } from "./_components/system-insights";
import { RecentActivity } from "./_components/recent-activity";
import { QuickActions } from "./_components/quick-actions";

export default function DashboardPage() {
  return (
    <DashboardProvider>
      <div className="space-y-[var(--spacing-gap)]">
        <DashboardStats />
        <SystemInsights />

        <div className="grid gap-[var(--spacing-gap)] md:grid-cols-7">
          <RecentActivity />
          <QuickActions />
        </div>
      </div>
    </DashboardProvider>
  );
}
