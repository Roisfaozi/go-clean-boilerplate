"use client";

import { Box, Layers } from "lucide-react";
import { Button } from "~/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/components/ui/tooltip";
import { useDensityStore } from "~/stores/use-density-store";
import { cn } from "~/lib/utils";

export function DensitySwitcher() {
  const { density, toggleDensity } = useDensityStore();

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            onClick={toggleDensity}
            className="w-9 px-0"
          >
            <div className="relative">
              {/* Animated Icon Transition */}
              <Box
                className={cn(
                  "absolute top-1/2 left-1/2 h-5 w-5 -translate-x-1/2 -translate-y-1/2 transition-all",
                  density === "compact"
                    ? "scale-0 rotate-90 opacity-0"
                    : "scale-100 rotate-0 opacity-100"
                )}
              />
              <Layers
                className={cn(
                  "absolute top-1/2 left-1/2 h-5 w-5 -translate-x-1/2 -translate-y-1/2 transition-all",
                  density === "comfort"
                    ? "scale-0 -rotate-90 opacity-0"
                    : "scale-100 rotate-0 opacity-100"
                )}
              />
            </div>
            <span className="sr-only">
              Switch to {density === "comfort" ? "Compact" : "Comfort"} mode
            </span>
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>
            {density === "comfort"
              ? "Switch to Compact (Enterprise)"
              : "Switch to Comfort (SaaS)"}
          </p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
