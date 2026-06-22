import { create } from "zustand";
import { persist } from "zustand/middleware";

type ThemeMode = "light" | "dark";
type DensityMode = "comfort" | "compact";

interface UIState {
  sidebarOpen: boolean;
  sidebarCollapsed: boolean;
  theme: ThemeMode;
  density: DensityMode;
  sidebarSectionOpen: Record<string, boolean>;
  toggleSidebar: () => void;
  toggleSidebarCollapse: () => void;
  setTheme: (theme: ThemeMode) => void;
  setDensity: (density: DensityMode) => void;
  setSidebarSectionOpen: (key: string, open: boolean) => void;
}

export const useUIStore = create<UIState>()(
  persist(
    (set, get) => ({
      sidebarOpen: true,
      sidebarCollapsed: false,
      theme: "light",
      density: "comfort",
      sidebarSectionOpen: {},
      toggleSidebar: () => set({ sidebarOpen: !get().sidebarOpen }),
      toggleSidebarCollapse: () =>
        set({ sidebarCollapsed: !get().sidebarCollapsed }),
      setTheme: (theme) => {
        document.documentElement.classList.toggle("dark", theme === "dark");
        set({ theme });
      },
      setDensity: (density) => {
        document.documentElement.classList.toggle(
          "density-compact",
          density === "compact",
        );
        set({ density });
      },
      setSidebarSectionOpen: (key, open) =>
        set({
          sidebarSectionOpen: { ...get().sidebarSectionOpen, [key]: open },
        }),
    }),
    { name: "nexus-ui" },
  ),
);
