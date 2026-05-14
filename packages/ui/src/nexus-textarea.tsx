import * as React from "react";
import { cn } from "./lib/utils";

const NexusTextarea = React.forwardRef<
  HTMLTextAreaElement,
  React.TextareaHTMLAttributes<HTMLTextAreaElement>
>(({ className, ...props }, ref) => {
  return (
    <textarea
      className={cn(
        "border-border bg-background text-body ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring duration-normal flex min-h-[80px] w-full rounded-md border px-[var(--input-padding-x)] py-[var(--input-padding-y)] transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      ref={ref}
      {...props}
    />
  );
});
NexusTextarea.displayName = "NexusTextarea";

export { NexusTextarea };
