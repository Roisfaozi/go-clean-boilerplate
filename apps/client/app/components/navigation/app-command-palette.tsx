import { useNavigate } from "react-router";
import {
  LayoutDashboard,
  Users,
  Shield,
  Lock,
  KeyRound,
  Building2,
  FolderKanban,
  FileText,
  Settings,
  Palette,
  Component,
} from "lucide-react";
import { CommandPalette, type CommandPaletteItem } from "./command-palette";
import { useState, useEffect } from "react";

export function AppCommandPalette() {
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((o) => !o);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  const items: CommandPaletteItem[] = [
    {
      id: "dashboard",
      label: "Dashboard",
      icon: LayoutDashboard,
      group: "Navigation",
      shortcut: "⌘1",
      onSelect: () => navigate("/"),
    },
    {
      id: "users",
      label: "Users",
      icon: Users,
      group: "Navigation",
      onSelect: () => navigate("/users"),
    },
    {
      id: "roles",
      label: "Roles",
      icon: Shield,
      group: "Navigation",
      onSelect: () => navigate("/roles"),
    },
    {
      id: "permissions",
      label: "Permissions",
      icon: Lock,
      group: "Navigation",
      onSelect: () => navigate("/permissions"),
    },
    {
      id: "access-rights",
      label: "Access Rights",
      icon: KeyRound,
      group: "Navigation",
      onSelect: () => navigate("/access-rights"),
    },
    {
      id: "organizations",
      label: "Organizations",
      icon: Building2,
      group: "Organization",
      onSelect: () => navigate("/organizations"),
    },
    {
      id: "projects",
      label: "Projects",
      icon: FolderKanban,
      group: "Organization",
      onSelect: () => navigate("/projects"),
    },
    {
      id: "audit-logs",
      label: "Audit Logs",
      icon: FileText,
      group: "System",
      onSelect: () => navigate("/audit-logs"),
    },
    {
      id: "settings",
      label: "Settings",
      icon: Settings,
      group: "System",
      onSelect: () => navigate("/settings"),
    },
    {
      id: "design-system",
      label: "Design System",
      icon: Palette,
      group: "System",
      onSelect: () => navigate("/design-system"),
    },
    {
      id: "components",
      label: "Components",
      icon: Component,
      group: "System",
      onSelect: () => navigate("/components"),
    },
  ];

  return (
    <CommandPalette
      items={items}
      open={open}
      onOpenChange={setOpen}
      placeholder="Search pages… (⌘K)"
    />
  );
}
