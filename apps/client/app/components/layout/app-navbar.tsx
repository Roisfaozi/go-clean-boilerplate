import { useUIStore, useAuthStore } from "@/stores";
import {
  NexusButton,
  NexusInput,
  Avatar,
  AvatarFallback,
  AvatarImage,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@casbin/ui";
import {
  Search,
  Sun,
  Moon,
  Menu,
  Maximize2,
  Minimize2,
  LogOut,
  User as UserIcon,
  Settings as SettingsIcon,
} from "lucide-react";
import { authApi } from "@/lib/api/auth";
import { useNavigate } from "react-router";
import { NotificationBell } from "@/components/realtime/notification-bell";
import { RealtimeIndicator } from "@/components/realtime/realtime-indicator";
import { PresenceAvatars } from "@/components/realtime/presence-avatars";

export function AppNavbar() {
  const { theme, setTheme, density, setDensity, toggleSidebarCollapse } =
    useUIStore();
  const { logout: clearStore, user } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await authApi.logout();
    } catch (err) {
      console.error("Logout failed", err);
    } finally {
      clearStore();
      navigate("/login");
    }
  };

  const userInitial = user?.username?.charAt(0).toUpperCase() || "A";

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
          onClick={() =>
            setDensity(density === "comfort" ? "compact" : "comfort")
          }
          title={
            density === "comfort" ? "Switch to compact" : "Switch to comfort"
          }
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
          {theme === "light" ? (
            <Moon className="h-4 w-4" />
          ) : (
            <Sun className="h-4 w-4" />
          )}
        </NexusButton>

        <NotificationBell />

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="ml-2 focus:outline-none">
              <Avatar className="h-8 w-8 cursor-pointer ring-offset-2 transition-transform hover:scale-105 active:scale-95">
                <AvatarImage src={user?.avatar_url} />
                <AvatarFallback className="bg-primary/20 text-primary text-xs font-bold">
                  {userInitial}
                </AvatarFallback>
              </Avatar>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel className="font-normal">
              <div className="flex flex-col space-y-1">
                <p className="text-sm leading-none font-medium">
                  {user?.username}
                </p>
                <p className="text-muted-foreground text-xs leading-none">
                  {user?.email}
                </p>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="cursor-pointer">
              <UserIcon className="mr-2 h-4 w-4" />
              <span>Profile</span>
            </DropdownMenuItem>
            <DropdownMenuItem className="cursor-pointer">
              <SettingsIcon className="mr-2 h-4 w-4" />
              <span>Settings</span>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className="text-destructive focus:text-destructive cursor-pointer"
              onClick={handleLogout}
            >
              <LogOut className="mr-2 h-4 w-4" />
              <span>Log out</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
