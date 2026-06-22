import { Suspense } from "react";
import { Skeleton } from "./skeleton";

interface LoadingBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  variant?: "page" | "card" | "inline";
}

function PageSkeleton() {
  return (
    <div className="animate-in fade-in space-y-6 duration-300">
      <div className="space-y-2">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-72" />
      </div>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-32 rounded-lg" />
        ))}
      </div>
      <Skeleton className="h-64 rounded-lg" />
    </div>
  );
}

function CardSkeleton() {
  return (
    <div className="bg-card border-border animate-in fade-in space-y-4 rounded-lg border p-6 duration-300">
      <div className="space-y-2">
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-3 w-48" />
      </div>
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-10 w-full" />
        ))}
      </div>
    </div>
  );
}

function InlineSkeleton() {
  return (
    <div className="animate-in fade-in flex items-center gap-3 py-2 duration-300">
      <Skeleton className="h-5 w-5 rounded" />
      <Skeleton className="h-4 flex-1" />
    </div>
  );
}

const fallbacks: Record<string, React.ReactNode> = {
  page: <PageSkeleton />,
  card: <CardSkeleton />,
  inline: <InlineSkeleton />,
};

export function LoadingBoundary({
  children,
  fallback,
  variant = "page",
}: LoadingBoundaryProps) {
  return (
    <Suspense fallback={fallback ?? fallbacks[variant]}>{children}</Suspense>
  );
}

export { PageSkeleton, CardSkeleton, InlineSkeleton };
