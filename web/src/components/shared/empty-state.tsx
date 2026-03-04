"use client";

import * as React from "react";
import { Icon } from "./icon";
import { Button } from "~/components/ui/button";
import { cn } from "~/lib/utils";
import { useDensity } from "./providers/density-provider";
import { motion, AnimatePresence } from "framer-motion";

export type EmptyStateCase =
  | "onboarding"
  | "search"
  | "filter"
  | "activity"
  | "users"
  | "roles"
  | "resources"
  | "security"
  | "error"
  | "generic";

export type EmptyStateVariant = "hero" | "embedded" | "minimal";

interface EmptyStateProps {
  case?: EmptyStateCase;
  variant?: EmptyStateVariant;
  title?: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
    icon?: string;
  };
  className?: string;
  searchTerm?: string;
}

const CASE_CONFIG: Record<
  EmptyStateCase,
  { icon: string; title: string; description: string; color: string }
> = {
  onboarding: {
    icon: "Rocket",
    title: "Welcome to NexusOS! 🎉",
    description:
      "Your admin dashboard is ready to be configured. Let's set up your first workspace.",
    color: "text-indigo-500",
  },
  search: {
    icon: "SearchX",
    title: "No results found",
    description:
      "We couldn't find what you were looking for. Try adjusting your keywords.",
    color: "text-muted-foreground",
  },
  filter: {
    icon: "FilterX",
    title: "No matches found",
    description:
      "Your current filters are too restrictive. Try clearing some to see more data.",
    color: "text-muted-foreground",
  },
  activity: {
    icon: "Inbox",
    title: "No activity yet",
    description:
      "System logs and user actions will appear here as they happen.",
    color: "text-muted-foreground/50",
  },
  users: {
    icon: "UserPlus",
    title: "No users found",
    description:
      "Start managing your team by adding or inviting your first member.",
    color: "text-primary",
  },
  roles: {
    icon: "ShieldAlert",
    title: "No roles defined",
    description:
      "Roles define what your users can do. Create one to start managing permissions.",
    color: "text-amber-500",
  },
  resources: {
    icon: "Database",
    title: "No resources found",
    description:
      "Add API endpoints and group them into resources to configure access.",
    color: "text-primary",
  },
  security: {
    icon: "Lock",
    title: "Access Denied",
    description:
      "You don't have the required permissions to view this section.",
    color: "text-destructive",
  },
  error: {
    icon: "AlertTriangle",
    title: "Unable to load data",
    description:
      "A connection error occurred. Please check your internet or try again.",
    color: "text-destructive",
  },
  generic: {
    icon: "FileQuestion",
    title: "Nothing here yet",
    description:
      "This section is currently empty. Check back later or add some data.",
    color: "text-muted-foreground",
  },
};

export function EmptyState({
  case: stateCase = "generic",
  variant = "embedded",
  title,
  description,
  action,
  className,
  searchTerm,
}: EmptyStateProps) {
  const { density } = useDensity();
  const isCompact = density === "compact";
  const config = CASE_CONFIG[stateCase];

  const displayTitle = searchTerm
    ? `No results for "${searchTerm}"`
    : (title ?? config.title);
  const displayDescription = description ?? config.description;

  // Render Logic based on Variant and Density
  if (variant === "minimal" || (isCompact && variant === "embedded")) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 5 }}
        animate={{ opacity: 1, y: 0 }}
        className={cn(
          "bg-muted/5 flex items-center gap-3 rounded-lg border border-dashed p-4",
          className
        )}
      >
        <Icon
          name={config.icon as any}
          size="sm"
          className={cn("shrink-0", config.color)}
        />
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-semibold">{displayTitle}</p>
        </div>
        {action && (
          <Button
            variant="ghost"
            size="sm"
            onClick={action.onClick}
            className="h-7 px-2 text-xs"
          >
            {action.label}
          </Button>
        )}
      </motion.div>
    );
  }

  return (
    <AnimatePresence mode="wait">
      <motion.div
        initial={{ opacity: 0, scale: 0.98 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.98 }}
        className={cn(
          "flex flex-col items-center justify-center text-center",
          variant === "hero"
            ? "px-6 py-24"
            : "bg-muted/5 rounded-xl border border-dashed px-4 py-12",
          className
        )}
      >
        <div
          className={cn(
            "bg-muted/10 ring-muted/5 mb-6 flex items-center justify-center rounded-full ring-8",
            variant === "hero" ? "h-24 w-24" : "h-16 w-16"
          )}
        >
          <Icon
            name={config.icon as any}
            className={cn(
              config.color,
              variant === "hero" ? "h-10 w-10" : "h-7 w-7"
            )}
          />
        </div>

        <div className="max-w-md space-y-2">
          <h3
            className={cn(
              "text-foreground font-bold tracking-tight",
              variant === "hero" ? "text-3xl" : "text-xl"
            )}
          >
            {displayTitle}
          </h3>
          <p className="text-muted-foreground text-sm leading-relaxed">
            {displayDescription}
          </p>
        </div>

        {action && (
          <div className="mt-8">
            <Button
              onClick={action.onClick}
              size={variant === "hero" ? "lg" : "default"}
              className={cn(
                "font-semibold transition-all hover:scale-105 active:scale-95",
                variant === "hero"
                  ? "shadow-primary/20 h-12 rounded-xl px-8 shadow-xl"
                  : ""
              )}
            >
              {action.icon && (
                <Icon name={action.icon as any} className="mr-2 h-4 w-4" />
              )}
              {action.label}
            </Button>
          </div>
        )}
      </motion.div>
    </AnimatePresence>
  );
}
