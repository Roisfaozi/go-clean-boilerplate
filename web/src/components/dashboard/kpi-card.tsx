"use client";

import { useDensity } from "~/components/shared/providers/density-provider";
import { cn } from "~/lib/utils";
import { Icon } from "~/components/shared/icon";
import { icons } from "lucide-react";

interface KPICardProps {
  title: string;
  value: string | number;
  trend?: string;
  trendUp?: boolean; // true for positive (green), false for negative (red)
  iconName: keyof typeof icons;
  description?: string;
}

export function KPICard({ title, value, trend, trendUp, iconName, description }: KPICardProps) {
  const { density } = useDensity();
  const isCompact = density === "compact";

  return (
    <div
      className={cn(
        "relative overflow-hidden transition-all duration-300",
        // Base styles
        "bg-card text-card-foreground",
        // Comfort Mode Styles
        !isCompact && "rounded-[var(--radius-xl)] p-6 shadow-md border-transparent hover:shadow-lg",
        // Compact Mode Styles
        isCompact && "rounded-[var(--radius-md)] p-3 border border-border shadow-none"
      )}
    >
      <div className={cn("flex justify-between items-start", isCompact && "items-center")}>
        <div className="space-y-1">
          <p className={cn("text-muted-foreground font-medium", isCompact ? "text-xs" : "text-sm")}>
            {title}
          </p>
          <div className="flex items-baseline gap-2">
            <h3 className={cn("font-bold tracking-tight", isCompact ? "text-xl" : "text-3xl")}>
              {value}
            </h3>
            {trend && (
              <span
                className={cn(
                  "text-xs font-medium",
                  trendUp ? "text-emerald-500" : "text-destructive"
                )}
              >
                {trend}
              </span>
            )}
          </div>
          {!isCompact && description && (
            <p className="text-xs text-muted-foreground pt-1">{description}</p>
          )}
        </div>

        {/* Icon Logic */}
        {!isCompact ? (
          <div className="p-3 bg-primary/10 rounded-full text-primary">
            <Icon name={iconName} className="h-6 w-6" />
          </div>
        ) : (
          <div className="text-muted-foreground/50">
             {/* In compact mode, maybe a smaller icon or just a sparkline placeholder */}
             <Icon name={iconName} className="h-4 w-4" />
          </div>
        )}
      </div>
      
      {/* Decorative background for Comfort mode */}
      {!isCompact && (
        <div className="absolute -right-4 -bottom-4 h-24 w-24 rounded-full bg-primary/5 blur-2xl" />
      )}
    </div>
  );
}
