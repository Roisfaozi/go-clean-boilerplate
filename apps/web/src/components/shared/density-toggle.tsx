"use client";

import { useDensityStore } from "~/stores/use-density-store";
import { Button } from "~/components/ui/button";
import { Layout, LayoutGrid } from "lucide-react";
import { useEffect } from "react";
import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "~/components/ui/tooltip";

export default function DensityToggle() {
  const { density, toggleDensity } = useDensityStore();

  useEffect(() => {
    document.documentElement.setAttribute("data-density", density);
  }, [density]);

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            onClick={toggleDensity}
            className="h-9 w-9 rounded-full"
          >
            {density === "comfort" ? (
              <Layout className="h-[1.2rem] w-[1.2rem]" />
            ) : (
              <LayoutGrid className="h-[1.2rem] w-[1.2rem]" />
            )}
            <span className="sr-only">Toggle density mode</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>Switch to {density === "comfort" ? "Compact" : "Comfort"} mode</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
