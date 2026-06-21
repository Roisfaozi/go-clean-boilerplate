import { useEffect, useState } from "react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@casbin/ui";
import type { LucideIcon } from "lucide-react";

export interface CommandPaletteItem {
  id: string;
  label: string;
  icon?: LucideIcon;
  shortcut?: string;
  group?: string;
  onSelect: () => void;
}

interface CommandPaletteProps {
  items: CommandPaletteItem[];
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  placeholder?: string;
}

export function CommandPalette({
  items,
  open: controlledOpen,
  onOpenChange,
  placeholder = "Type a command or search…",
}: CommandPaletteProps) {
  const [internalOpen, setInternalOpen] = useState(false);
  const isOpen = controlledOpen ?? internalOpen;
  const setOpen = onOpenChange ?? setInternalOpen;

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen(!isOpen);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, [isOpen, setOpen]);

  const groups = items.reduce<Record<string, CommandPaletteItem[]>>(
    (acc, item) => {
      const group = item.group || "Actions";
      if (!acc[group]) acc[group] = [];
      acc[group].push(item);
      return acc;
    },
    {},
  );

  return (
    <CommandDialog open={isOpen} onOpenChange={setOpen}>
      <CommandInput placeholder={placeholder} />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>
        {Object.entries(groups).map(([group, groupItems], i) => (
          <div key={group}>
            {i > 0 && <CommandSeparator />}
            <CommandGroup heading={group}>
              {groupItems.map((item) => (
                <CommandItem
                  key={item.id}
                  onSelect={() => {
                    item.onSelect();
                    setOpen(false);
                  }}
                >
                  {item.icon && <item.icon className="mr-2 h-4 w-4" />}
                  <span>{item.label}</span>
                  {item.shortcut && (
                    <CommandShortcut>{item.shortcut}</CommandShortcut>
                  )}
                </CommandItem>
              ))}
            </CommandGroup>
          </div>
        ))}
      </CommandList>
    </CommandDialog>
  );
}
