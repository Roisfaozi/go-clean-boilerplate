"use client";

import * as React from "react";
import { ScrollArea } from "@base-ui/react/scroll-area";

import { cn } from "./lib/utils";

const ScrollAreaRoot = React.forwardRef<
  React.ComponentRef<typeof ScrollArea.Root>,
  React.ComponentProps<typeof ScrollArea.Root>
>(({ className, children, ...props }, ref) => (
  <ScrollArea.Root
    ref={ref}
    className={cn("relative overflow-hidden", className)}
    {...props}
  >
    <ScrollArea.Viewport className="h-full w-full rounded-[inherit]">
      {children}
    </ScrollArea.Viewport>
    <ScrollBar />
    <ScrollArea.Corner />
  </ScrollArea.Root>
));
ScrollAreaRoot.displayName = "ScrollArea";

const ScrollBar = React.forwardRef<
  React.ComponentRef<typeof ScrollArea.Scrollbar>,
  React.ComponentProps<typeof ScrollArea.Scrollbar> & { orientation?: "vertical" | "horizontal" }
>(({ className, orientation = "vertical", ...props }, ref) => (
  <ScrollArea.Scrollbar
    ref={ref}
    orientation={orientation}
    className={cn(
      "flex touch-none transition-colors select-none",
      orientation === "vertical" && "h-full w-2.5 border-l border-l-transparent p-px",
      orientation === "horizontal" && "h-2.5 flex-col border-t border-t-transparent p-px",
      className,
    )}
    {...props}
  >
    <ScrollArea.Thumb className="bg-border relative flex-1 rounded-full" />
  </ScrollArea.Scrollbar>
));
ScrollBar.displayName = "ScrollBar";

export { ScrollAreaRoot as ScrollArea, ScrollBar };
