"use client";

import { DashboardHeader } from "~/components/layout/dashboard/header";
import { Sidebar } from "~/components/layout/sidebar";
import { usePresence } from "~/hooks/use-presence";
import { DashboardShellProvider } from "./_components/dashboard-shell-context";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  usePresence();

  return (
    <DashboardShellProvider>
      <div className="bg-background flex min-h-screen">
        {/* Sidebar */}
        <Sidebar className="z-40 hidden md:flex" />

        {/* Main Area */}
        <div className="flex min-h-screen flex-1 flex-col transition-all">
          <DashboardHeader />

          <main className="flex-1 overflow-y-auto p-[var(--layout-padding)]">
            {children}
          </main>
        </div>
      </div>
    </DashboardShellProvider>
  );
}
