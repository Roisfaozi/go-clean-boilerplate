import { useState } from "react";
import { cn } from "@casbin/ui";
import { useOrganizationStore } from "@/stores/organization-store";
import { Check, ChevronsUpDown, Plus } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@casbin/ui";
import type { Organization } from "@/lib/api/types";

const mockOrgs: Organization[] = [
  {
    id: "1",
    name: "Acme Corp",
    slug: "acme",
    owner_id: "u1",
    status: "active",
  },
  {
    id: "2",
    name: "Globex Inc",
    slug: "globex",
    owner_id: "u2",
    status: "active",
  },
  {
    id: "3",
    name: "Initech",
    slug: "initech",
    owner_id: "u1",
    status: "active",
  },
];

interface OrganizationSwitcherProps {
  collapsed?: boolean;
}

export function OrganizationSwitcher({ collapsed }: OrganizationSwitcherProps) {
  const [open, setOpen] = useState(false);
  const { activeOrganization, setActiveOrganization } = useOrganizationStore();

  const current = activeOrganization ?? mockOrgs[0];

  const handleSelect = (org: Organization) => {
    setActiveOrganization(org);
    setOpen(false);
  };

  if (collapsed) {
    return (
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <button className="bg-primary/10 text-primary hover:bg-primary/20 mx-auto flex h-10 w-10 items-center justify-center rounded-md text-sm font-bold transition-colors">
            {current.name.charAt(0)}
          </button>
        </PopoverTrigger>
        <PopoverContent side="right" align="start" className="w-56 p-1">
          {mockOrgs.map((org) => (
            <button
              key={org.id}
              onClick={() => handleSelect(org)}
              className={cn(
                "flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors",
                current.id === org.id
                  ? "bg-primary/10 text-primary"
                  : "hover:bg-muted",
              )}
            >
              <div className="bg-primary/10 text-primary flex h-6 w-6 shrink-0 items-center justify-center rounded text-xs font-bold">
                {org.name.charAt(0)}
              </div>
              <span className="truncate">{org.name}</span>
              {current.id === org.id && (
                <Check className="ml-auto h-4 w-4 shrink-0" />
              )}
            </button>
          ))}
        </PopoverContent>
      </Popover>
    );
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button className="hover:bg-sidebar-accent flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left transition-colors">
          <div className="bg-primary/10 text-primary flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-sm font-bold">
            {current.name.charAt(0)}
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-sidebar-foreground truncate text-sm font-medium">
              {current.name}
            </p>
            <p className="text-muted-foreground text-[11px]">Organization</p>
          </div>
          <ChevronsUpDown className="text-muted-foreground h-4 w-4 shrink-0" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-60 p-1">
        <p className="text-muted-foreground px-3 py-1.5 text-[11px] font-semibold tracking-wider uppercase">
          Organizations
        </p>
        {mockOrgs.map((org) => (
          <button
            key={org.id}
            onClick={() => handleSelect(org)}
            className={cn(
              "flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors",
              current.id === org.id
                ? "bg-primary/10 text-primary"
                : "hover:bg-muted",
            )}
          >
            <div className="bg-primary/10 text-primary flex h-7 w-7 shrink-0 items-center justify-center rounded text-xs font-bold">
              {org.name.charAt(0)}
            </div>
            <span className="flex-1 truncate text-left">{org.name}</span>
            {current.id === org.id && <Check className="h-4 w-4 shrink-0" />}
          </button>
        ))}
        <div className="border-border mt-1 border-t pt-1">
          <button className="text-muted-foreground hover:bg-muted flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors">
            <Plus className="h-4 w-4" />
            Create Organization
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
