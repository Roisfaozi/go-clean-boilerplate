"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { buttonVariants } from "~/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/components/ui/tooltip";
import { cn } from "~/lib/utils";
import { OrganizationSwitcher } from "../dashboard/organization-switcher";
import { Icon } from "../shared/icon";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { memo } from "react";

// Define Navigation Items (Could be moved to a separate config file)
const navItems = [
  {
    title: "Dashboard",
    href: "/dashboard",
    iconName: "LayoutDashboard" as const,
  },
  {
    title: "Projects",
    href: "/dashboard/projects",
    iconName: "Folder" as const,
  },
  {
    title: "Users",
    href: "/dashboard/users",
    iconName: "UserSearch" as const,
  },
  {
    title: "Team Members",
    href: "/dashboard/organization/members",
    iconName: "Users" as const,
  },
  {
    title: "Org Settings",
    href: "/dashboard/organization/settings",
    iconName: "Building" as const,
  },
  {
    title: "Roles",
    href: "/dashboard/roles",
    iconName: "Shield" as const,
  },
  {
    title: "Access Matrix",
    href: "/dashboard/access",
    iconName: "Grid3X3" as const,
  },
  {
    title: "Access Rights",
    href: "/dashboard/access-rights",
    iconName: "Key" as const,
  },
  {
    title: "Audit Logs",
    href: "/dashboard/audit",
    iconName: "FileText" as const,
  },
  {
    title: "Settings",
    href: "/dashboard/settings",
    iconName: "Settings" as const,
  },
];

export const Sidebar = memo(function Sidebar({
  className,
}: {
  className?: string;
}) {
  const pathname = usePathname();
  const { currentOrganization } = useDashboardShell();

  return (
    <aside
      className={cn(
        "bg-background flex flex-col border-r transition-all duration-300",
        "sticky top-0 h-screen w-[var(--sidebar-width)]",
        className
      )}
    >
      {/* Header / Logo + Switcher */}
      <div className="flex h-[var(--navbar-height)] items-center gap-2 border-b px-3">
        <Link
          href="/"
          className="flex shrink-0 items-center gap-2 overflow-hidden"
        >
          <Icon name="Command" size="md" className="text-primary" />
        </Link>
        <OrganizationSwitcher />
      </div>

      {/* Navigation */}
      <nav className="flex flex-1 flex-col gap-1 overflow-y-auto p-2">
        {navItems.map((item) => {
          const isActive =
            pathname === item.href || pathname.startsWith(`${item.href}/`);

          return (
            <TooltipProvider key={item.href}>
              <Tooltip delayDuration={0}>
                <TooltipTrigger asChild>
                  <Link
                    href={item.href}
                    className={cn(
                      buttonVariants({
                        variant: isActive ? "secondary" : "ghost",
                        size: "default",
                      }),
                      "w-full justify-start overflow-hidden",
                      isActive &&
                        "bg-primary/10 text-primary hover:bg-primary/20",
                      "[data-density=compact]:justify-center [data-density=compact]:px-0"
                    )}
                  >
                    <Icon
                      name={item.iconName as any}
                      className={cn(isActive && "text-primary")}
                    />
                    <span className="ml-3 truncate [data-density=compact]:hidden">
                      {item.title}
                    </span>
                  </Link>
                </TooltipTrigger>
                <TooltipContent
                  side="right"
                  className="hidden [data-density=compact]:block"
                >
                  {item.title}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          );
        })}
      </nav>

      {/* Footer / User Info Context (Optional) */}
      <div className="border-t p-4 [data-density=compact]:p-2">
        <div className="text-muted-foreground flex flex-col gap-1 text-[10px] font-bold tracking-wider uppercase [data-density=compact]:hidden">
          <span>Active Org</span>
          <span className="text-primary truncate">
            {currentOrganization?.name || "None"}
          </span>
        </div>
      </div>
    </aside>
  );
});
