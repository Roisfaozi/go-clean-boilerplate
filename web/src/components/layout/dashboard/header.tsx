"use client";

import { GlobalSearch } from "~/components/shared/global-search";
import { DensitySwitcher } from "~/components/shared/density-switcher";
import ThemeToggle from "~/components/shared/theme-toggle";
import LocaleToggler from "~/components/shared/locale-toggler";
import { UserNav } from "~/components/dashboard/user-nav"; // Need to create this
import { SidebarTrigger } from "~/components/ui/sidebar"; // From shadcn/sidebar if available? Or custom.
import { Separator } from "~/components/ui/separator";
import { cn } from "~/lib/utils";

export function DashboardHeader() {
  return (
    <header className={cn(
      "sticky top-0 z-30 flex items-center gap-4 border-b bg-background px-6 transition-all",
      // Density: Comfort 80px, Compact 56px
      "h-[var(--navbar-height)]"
    )}>
      {/* Left: Search & Trigger */}
      <div className="flex items-center gap-4 flex-1">
        <GlobalSearch />
      </div>

      {/* Right: Actions */}
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-1">
          <DensitySwitcher />
          <ThemeToggle />
          <LocaleToggler />
        </div>
        <Separator orientation="vertical" className="h-6" />
        <UserNav />
      </div>
    </header>
  );
}
