"use client";

import * as React from "react";
import { useDensityStore } from "~/stores/use-density-store";

export function DensityProvider({ children }: { children: React.ReactNode }) {
  const density = useDensityStore((state) => state.density);
  const [mounted, setMounted] = React.useState(false);

  React.useEffect(() => {
    setMounted(true);
  }, []);

  React.useEffect(() => {
    if (mounted) {
      document.documentElement.setAttribute("data-density", density);
    }
  }, [density, mounted]);

  if (!mounted) {
    return <>{children}</>;
  }

  return <>{children}</>;
}

export function useDensity() {
  const density = useDensityStore((state) => state.density);
  const toggleDensity = useDensityStore((state) => state.toggleDensity);
  const setDensity = useDensityStore((state) => state.setDensity);

  return {
    density,
    toggleDensity,
    setDensity,
  };
}
