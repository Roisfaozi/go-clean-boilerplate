import { create } from "zustand";
import { persist } from "zustand/middleware";

type Density = "comfort" | "compact";

interface DensityState {
  density: Density;
  toggleDensity: () => void;
  setDensity: (density: Density) => void;
}

export const useDensityStore = create<DensityState>()(
  persist(
    (set) => ({
      density: "comfort",
      toggleDensity: () =>
        set((state) => ({
          density: state.density === "comfort" ? "compact" : "comfort",
        })),
      setDensity: (density) => set({ density }),
    }),
    {
      name: "nexus-density-storage",
    }
  )
);
