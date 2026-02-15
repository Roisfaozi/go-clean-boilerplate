"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { buttonVariants } from "~/components/ui/button";
import { cn } from "~/lib/utils";
import Icons from "../shared/icons";
import { Icon } from "../shared/icon"; // Use the density-aware icon
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/components/ui/tooltip";
import { OrganizationSwitcher } from "../dashboard/organization-switcher";

// Define Navigation Items
const navItems = [
  {
    title: "Dashboard",
    href: "/dashboard",
    iconName: "LayoutDashboard" as const,
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

export function Sidebar({ className }: { className?: string }) {
  const pathname = usePathname();

  return (
    <aside
      className={cn(
        "bg-background flex flex-col border-r transition-all duration-300",
        "sticky top-0 h-screen w-[var(--sidebar-width)]",
        className
      )}
    >
      {/* Header / Logo + Switcher */}
      <div className="flex h-[var(--navbar-height)] items-center border-b px-3 gap-2">
        <Link href="/" className="flex items-center gap-2 overflow-hidden shrink-0">
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
                      // Compact Mode: Center icon, hide text
                      "[data-density=compact]:justify-center [data-density=compact]:px-0"
                    )}
                  >
                    {/* Icon */}
                    {/* We need to map iconName to the Icon component properly.
                        Since 'Icon' takes 'name' prop which is LucideKeys.
                        I'll use a dynamic mapping or just import specific icons if needed.
                        For now, assuming Icon component handles string names correctly.
                    */}
                    <Icon
                      name={item.iconName as any}
                      className={cn(isActive && "text-primary")}
                    />

                    {/* Label */}
                    <span className="ml-3 truncate [data-density=compact]:hidden">
                      {item.title}
                    </span>
                  </Link>
                </TooltipTrigger>
                {/* Tooltip only in Compact Mode */}
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

      {/* Footer / User Profile (Mini) */}
      <div className="border-t p-4 [data-density=compact]:p-2">
        {/* Could put user profile here or just logout */}
      </div>
    </aside>
  );
}
