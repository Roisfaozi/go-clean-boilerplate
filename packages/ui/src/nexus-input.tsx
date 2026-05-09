import * as React from "react";
import { cn } from "./lib/utils";

const NexusInput = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          "flex h-input w-full rounded-md border border-border bg-background px-[var(--input-padding-x)] py-[var(--input-padding-y)] text-body ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-colors duration-normal",
          className
        )}
        ref={ref}
        {...props}
      />
    );
  }
);
NexusInput.displayName = "NexusInput";

export { NexusInput };
