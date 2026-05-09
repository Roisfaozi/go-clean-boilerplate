import { useUIStore } from "@/stores";
import { NexusButton } from "@casbin/ui";
import { Search, Sun, Moon, Menu, Maximize2, Minimize2 } from "lucide-react";
import { NexusInput } from "@casbin/ui";
import { NotificationBell } from "@/components/realtime/notification-bell";
import { RealtimeIndicator } from "@/components/realtime/realtime-indicator";
import { PresenceAvatars } from "@/components/realtime/presence-avatars";

export function AppNavbar() {
  const { theme, setTheme, density, setDensity, toggleSidebarCollapse } = useUIStore();

  return (
    <header className="h-navbar border-border bg-background px-layout sticky top-0 z-10 flex items-center justify-between border-b">
      <div className="flex flex-1 items-center gap-3">
        <button
          onClick={toggleSidebarCollapse}
          className="hover:bg-surface-hover text-muted-foreground rounded-md p-2 lg:hidden"
        >
          <Menu className="h-5 w-5" />
        </button>
        <div className="relative hidden w-full max-w-md sm:block">
          <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <NexusInput placeholder="Search..." className="h-9 pl-10" />
        </div>
      </div>

      <div className="flex items-center gap-2">
        {/* Presence avatars (hidden on small screens) */}
        <div className="mr-2 hidden lg:block">
          <PresenceAvatars max={3} size="sm" showCount={false} />
        </div>

        <RealtimeIndicator showLabel className="mr-1" />

        <NexusButton
          variant="ghost"
          size="icon"
          onClick={() => setDensity(density === "comfort" ? "compact" : "comfort")}
          title={density === "comfort" ? "Switch to compact" : "Switch to comfort"}
        >
          {density === "comfort" ? (
            <Minimize2 className="h-4 w-4" />
          ) : (
            <Maximize2 className="h-4 w-4" />
          )}
        </NexusButton>
        <NexusButton
          variant="ghost"
          size="icon"
          onClick={() => setTheme(theme === "light" ? "dark" : "light")}
        >
          {theme === "light" ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
        </NexusButton>

        <NotificationBell />

        <div className="bg-primary/20 text-small text-primary ml-2 flex h-8 w-8 items-center justify-center rounded-full font-semibold">
          A
        </div>
      </div>
    </header>
  );
}
