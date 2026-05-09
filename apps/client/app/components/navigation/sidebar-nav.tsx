import { cn } from "@casbin/ui";
import { ChevronDown } from "lucide-react";
import { useState } from "react";
import { Link, useLocation } from "react-router";

/* ── SidebarGroup ── */
interface SidebarGroupProps {
  label?: string;
  children: React.ReactNode;
  className?: string;
}

export function SidebarGroup({ label, children, className }: SidebarGroupProps) {
  return (
    <div className={cn("space-y-1", className)}>
      {label && (
        <p className="text-caption font-semibold uppercase tracking-wider text-muted-foreground px-3 py-2">
          {label}
        </p>
      )}
      {children}
    </div>
  );
}

/* ── SidebarItem ── */
interface SidebarItemProps {
  label: string;
  href: string;
  icon?: React.ReactNode;
  badge?: React.ReactNode;
  collapsed?: boolean;
  className?: string;
}

export function SidebarItem({ label, href, icon, badge, collapsed, className }: SidebarItemProps) {
  const location = useLocation();
  const isActive = location.pathname === href;

  return (
    <Link
      to={href}
      className={cn(
        "flex items-center gap-3 px-3 py-2.5 rounded-md text-body font-medium transition-colors duration-normal",
        isActive
          ? "bg-primary/10 text-primary"
          : "text-muted-foreground hover:bg-surface-hover hover:text-foreground",
        className
      )}
    >
      {icon && <span className="shrink-0">{icon}</span>}
      {!collapsed && <span className="flex-1 truncate">{label}</span>}
      {!collapsed && badge}
    </Link>
  );
}

/* ── SidebarCollapsible ── */
interface SidebarCollapsibleProps {
  label: string;
  icon?: React.ReactNode;
  children: React.ReactNode;
  defaultOpen?: boolean;
  collapsed?: boolean;
}

export function SidebarCollapsible({ label, icon, children, defaultOpen = false, collapsed }: SidebarCollapsibleProps) {
  const [open, setOpen] = useState(defaultOpen);

  if (collapsed) {
    return <>{children}</>;
  }

  return (
    <div>
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-3 w-full px-3 py-2.5 rounded-md text-body font-medium text-muted-foreground hover:bg-surface-hover hover:text-foreground transition-colors duration-normal"
      >
        {icon && <span className="shrink-0">{icon}</span>}
        <span className="flex-1 text-left truncate">{label}</span>
        <ChevronDown
          className={cn("h-4 w-4 shrink-0 transition-transform duration-normal", open && "rotate-180")}
        />
      </button>
      {open && <div className="ml-6 mt-1 space-y-0.5">{children}</div>}
    </div>
  );
}
