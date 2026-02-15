"use client";

import { GlobalSearch } from "~/components/shared/global-search";
import { DashboardBreadcrumbs } from "./breadcrumbs";
import { DensitySwitcher } from "~/components/shared/density-switcher";
import { NotificationCenter } from "~/components/dashboard/notification-center";
import ThemeToggle from "~/components/shared/theme-toggle";
import LocaleToggler from "~/components/shared/locale-toggler";
import { UserNav } from "~/components/dashboard/user-nav";
import { SidebarTrigger } from "~/components/ui/sidebar";
import { Separator } from "~/components/ui/separator";
import { PresenceAvatarStack } from "~/components/dashboard/presence-avatar-stack";
import { cn } from "~/lib/utils";

export function DashboardHeader() {
  return (
    <header
      className={cn(
        "bg-background sticky top-0 z-30 flex items-center gap-4 border-b px-6 transition-all",
        // Density: Comfort 80px, Compact 56px
        "h-[var(--navbar-height)]"
      )}
    >
      {/* Left: Search & Trigger */}
      <div className="flex flex-1 items-center gap-4">
        <GlobalSearch />
        <Separator orientation="vertical" className="h-6 hidden md:block" />
        <DashboardBreadcrumbs />
      </div>

      {/* Right: Actions */}
      <div className="flex items-center gap-3">
        <PresenceAvatarStack className="hidden lg:flex" />
        <Separator orientation="vertical" className="h-6 hidden lg:block" />
        <div className="flex items-center gap-1">
          <NotificationCenter />
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
