"use client";

import { Skeleton } from "~/components/ui/skeleton";
import { cn } from "~/lib/utils";
import { Card, CardContent, CardHeader } from "~/components/ui/card";

/**
 * TableSkeleton - Standard pattern for data grids (Users, Audit, Members)
 */
export function TableSkeleton({ rows = 5, columns = 5 }: { rows?: number; columns?: number }) {
  return (
    <div className="w-full space-y-4">
      <div className="rounded-md border overflow-hidden">
        <div className="bg-muted/50 border-b p-4">
          <div className="flex gap-4">
            {Array.from({ length: columns }).map((_, i) => (
              <Skeleton key={i} className="h-4 flex-1" />
            ))}
          </div>
        </div>
        <div className="divide-y">
          {Array.from({ length: rows }).map((_, i) => (
            <div key={i} className="p-4 flex gap-4 items-center">
              {Array.from({ length: columns }).map((_, j) => (
                <Skeleton key={j} className={cn("h-4 flex-1", j === 0 && "h-8 w-8 rounded-full flex-none")} />
              ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

/**
 * CardGridSkeleton - Pattern for Role cards or Project grid
 */
export function CardGridSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {Array.from({ length: count }).map((_, i) => (
        <Card key={i} className="overflow-hidden">
          <CardHeader className="space-y-2">
            <Skeleton className="h-5 w-1/2" />
            <Skeleton className="h-4 w-3/4" />
          </CardHeader>
          <CardContent className="space-y-4">
            <Skeleton className="h-20 w-full" />
            <div className="flex justify-between items-center">
              <Skeleton className="h-8 w-20" />
              <Skeleton className="h-8 w-8 rounded-full" />
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

/**
 * FormSkeleton - Pattern for Modals and Settings forms
 */
export function FormSkeleton({ fields = 3 }: { fields?: number }) {
  return (
    <div className="space-y-6 py-4">
      {Array.from({ length: fields }).map((_, i) => (
        <div key={i} className="space-y-2">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-10 w-full" />
        </div>
      ))}
      <div className="flex justify-end gap-3 pt-4">
        <Skeleton className="h-10 w-24" />
        <Skeleton className="h-10 w-32" />
      </div>
    </div>
  );
}

/**
 * KPISkeleton - Specifically for Dashboard metric cards
 */
export function KPISkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <Card key={i} className="p-6">
          <div className="flex items-center justify-between space-y-0 pb-2">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-8 w-8 rounded-full" />
          </div>
          <div className="space-y-2">
            <Skeleton className="h-8 w-16" />
            <Skeleton className="h-3 w-32" />
          </div>
        </Card>
      ))}
    </div>
  );
}

/**
 * StatsChartSkeleton - For Activity charts
 */
export function StatsChartSkeleton() {
  const barHeights = [40, 70, 45, 90, 65, 50, 80, 35, 60, 75, 40, 55];

  return (
    <Card className="w-full">
      <CardHeader>
        <Skeleton className="h-6 w-48" />
        <Skeleton className="h-4 w-64" />
      </CardHeader>
      <CardContent>
        <div className="h-[250px] flex items-end gap-2 px-2">
          {barHeights.map((height, i) => (
            <Skeleton 
              key={i} 
              className="flex-1 rounded-t-sm" 
              style={{ height: `${height}%` }} 
            />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
