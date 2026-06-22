import { cn } from "@casbin/ui";
import { TrendingUp, TrendingDown, type LucideIcon } from "lucide-react";

interface StatCardProps {
  title: string;
  value: string | number;
  trend?: { value: number; label?: string };
  icon?: LucideIcon;
  className?: string;
}

export function StatCard({
  title,
  value,
  trend,
  icon: Icon,
  className,
}: StatCardProps) {
  const isPositive = trend && trend.value >= 0;
  return (
    <div
      className={cn(
        "bg-card border-border p-card-pad rounded-lg border shadow-sm",
        className,
      )}
    >
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <p className="text-small text-muted-foreground">{title}</p>
          <p className="text-h1 font-bold">{value}</p>
        </div>
        {Icon && (
          <div className="bg-primary/10 rounded-md p-2.5">
            <Icon className="text-primary h-5 w-5" />
          </div>
        )}
      </div>
      {trend && (
        <div className="mt-3 flex items-center gap-1">
          {isPositive ? (
            <TrendingUp className="text-success h-4 w-4" />
          ) : (
            <TrendingDown className="text-danger h-4 w-4" />
          )}
          <span
            className={cn(
              "text-caption font-medium",
              isPositive ? "text-success" : "text-danger",
            )}
          >
            {isPositive ? "+" : ""}
            {trend.value}%
          </span>
          {trend.label && (
            <span className="text-caption text-muted-foreground">
              {trend.label}
            </span>
          )}
        </div>
      )}
    </div>
  );
}
