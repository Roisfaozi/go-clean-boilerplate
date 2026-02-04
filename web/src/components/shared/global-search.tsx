"use client";

import * as React from "react";
import {
  Calculator,
  Calendar,
  CreditCard,
  Settings,
  Smile,
  User,
  Search,
  LayoutDashboard,
  Shield,
  FileText,
} from "lucide-react";

import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "~/components/ui/command";
import { Button } from "~/components/ui/button";
import { cn } from "~/lib/utils";
import { useDensityStore } from "~/stores/use-density-store";

export function GlobalSearch() {
  const [open, setOpen] = React.useState(false);
  const { density } = useDensityStore();

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };

    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  return (
    <>
      <Button
        variant="outline"
        className={cn(
          "text-muted-foreground hover:bg-accent relative justify-start transition-all",
          // Comfort: Full width input-like
          "w-full md:w-64 lg:w-80",
          // Compact: Icon only or smaller width
          "[data-density=compact]:w-10 [data-density=compact]:justify-center [data-density=compact]:px-0"
        )}
        onClick={() => setOpen(true)}
      >
        <Search className="h-4 w-4 shrink-0 [data-density=comfort]:mr-2" />
        <span className="inline-flex [data-density=compact]:hidden">
          Search...
        </span>
        <kbd className="bg-muted pointer-events-none absolute top-1/2 right-2 hidden h-5 -translate-y-1/2 items-center gap-1 rounded border px-1.5 font-mono text-[10px] font-medium opacity-100 select-none md:flex [data-density=compact]:hidden">
          <span className="text-xs">⌘</span>K
        </kbd>
      </Button>

      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Type a command or search..." />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
          <CommandGroup heading="Navigation">
            <CommandItem>
              <LayoutDashboard className="mr-2 h-4 w-4" />
              <span>Dashboard</span>
            </CommandItem>
            <CommandItem>
              <User className="mr-2 h-4 w-4" />
              <span>Users</span>
            </CommandItem>
            <CommandItem>
              <Shield className="mr-2 h-4 w-4" />
              <span>Roles & Access</span>
            </CommandItem>
            <CommandItem>
              <FileText className="mr-2 h-4 w-4" />
              <span>Audit Logs</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Actions">
            <CommandItem>
              <User className="mr-2 h-4 w-4" />
              <span>Add User</span>
              <CommandShortcut>⌘U</CommandShortcut>
            </CommandItem>
            <CommandItem>
              <CreditCard className="mr-2 h-4 w-4" />
              <span>Billing</span>
              <CommandShortcut>⌘B</CommandShortcut>
            </CommandItem>
            <CommandItem>
              <Settings className="mr-2 h-4 w-4" />
              <span>Settings</span>
              <CommandShortcut>⌘S</CommandShortcut>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </>
  );
}
