import { DashboardHeader } from "~/components/layout/dashboard/header";
import { Sidebar } from "~/components/layout/sidebar";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen bg-background">
      {/* Sidebar */}
      <Sidebar className="hidden md:flex z-40" />

      {/* Main Area */}
      <div className="flex-1 flex flex-col min-h-screen transition-all">
        <DashboardHeader />
        
        <main className="flex-1 p-[var(--layout-padding)] overflow-y-auto">
          {children}
        </main>
      </div>
    </div>
  );
}