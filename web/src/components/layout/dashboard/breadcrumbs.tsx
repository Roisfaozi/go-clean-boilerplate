"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ChevronRight, Home } from "lucide-react";
import { cn } from "~/lib/utils";

export function DashboardBreadcrumbs() {
  const pathname = usePathname();
  const paths = pathname.split("/").filter((p) => p && p !== "en" && p !== "fr"); // Filter locale and empty

  // If we are just on /dashboard, show nothing or just Home
  if (paths.length === 0) return null;

  return (
    <nav className="flex items-center text-sm font-medium text-muted-foreground ml-2 hidden md:flex">
      <div className="flex items-center">
        <Link 
            href="/dashboard" 
            className="hover:text-foreground transition-colors flex items-center gap-1"
        >
          <Home className="h-3.5 w-3.5" />
          <span className="sr-only">Home</span>
        </Link>
      </div>
      
      {paths.map((path, index) => {
        const href = `/${paths.slice(0, index + 1).join("/")}`;
        const isLast = index === paths.length - 1;
        const label = path.charAt(0).toUpperCase() + path.slice(1).replace(/-/g, " ");

        if (path === "dashboard" && index === 0) return null;

        return (
          <div key={path} className="flex items-center">
            <ChevronRight className="h-4 w-4 mx-1 opacity-50 shrink-0" />
            {isLast ? (
              <span className="text-foreground font-semibold truncate max-w-[150px]">
                {label}
              </span>
            ) : (
              <Link
                href={href}
                className="hover:text-foreground transition-colors truncate max-w-[150px]"
              >
                {label}
              </Link>
            )}
          </div>
        );
      })}
    </nav>
  );
}
