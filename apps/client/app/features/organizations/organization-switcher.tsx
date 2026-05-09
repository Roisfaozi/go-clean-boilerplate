import { useState } from "react";
import { cn } from "@casbin/ui";
import { useOrganizationStore } from "@/stores/organization-store";
import { Building2, Check, ChevronsUpDown, Plus } from "lucide-react";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@casbin/ui";
import {  NexusButton  } from "@casbin/ui";
import type { Organization } from "@/lib/api/types";

const mockOrgs: Organization[] = [
  { id: "1", name: "Acme Corp", slug: "acme", owner_id: "u1", status: "active" },
  { id: "2", name: "Globex Inc", slug: "globex", owner_id: "u2", status: "active" },
  { id: "3", name: "Initech", slug: "initech", owner_id: "u1", status: "active" },
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
          <button className="flex items-center justify-center h-10 w-10 mx-auto rounded-md bg-primary/10 text-primary font-bold text-sm hover:bg-primary/20 transition-colors">
            {current.name.charAt(0)}
          </button>
        </PopoverTrigger>
        <PopoverContent side="right" align="start" className="w-56 p-1">
          {mockOrgs.map((org) => (
            <button
              key={org.id}
              onClick={() => handleSelect(org)}
              className={cn(
                "flex items-center gap-2 w-full px-3 py-2 rounded-md text-sm transition-colors",
                current.id === org.id ? "bg-primary/10 text-primary" : "hover:bg-muted"
              )}
            >
              <div className="h-6 w-6 rounded bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">
                {org.name.charAt(0)}
              </div>
              <span className="truncate">{org.name}</span>
              {current.id === org.id && <Check className="h-4 w-4 ml-auto shrink-0" />}
            </button>
          ))}
        </PopoverContent>
      </Popover>
    );
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button className="flex items-center gap-3 w-full px-3 py-2.5 rounded-md hover:bg-sidebar-accent transition-colors text-left">
          <div className="h-8 w-8 rounded-md bg-primary/10 flex items-center justify-center text-sm font-bold text-primary shrink-0">
            {current.name.charAt(0)}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-sidebar-foreground truncate">{current.name}</p>
            <p className="text-[11px] text-muted-foreground">Organization</p>
          </div>
          <ChevronsUpDown className="h-4 w-4 text-muted-foreground shrink-0" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-60 p-1">
        <p className="px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
          Organizations
        </p>
        {mockOrgs.map((org) => (
          <button
            key={org.id}
            onClick={() => handleSelect(org)}
            className={cn(
              "flex items-center gap-2 w-full px-3 py-2 rounded-md text-sm transition-colors",
              current.id === org.id ? "bg-primary/10 text-primary" : "hover:bg-muted"
            )}
          >
            <div className="h-7 w-7 rounded bg-primary/10 flex items-center justify-center text-xs font-bold text-primary shrink-0">
              {org.name.charAt(0)}
            </div>
            <span className="truncate flex-1 text-left">{org.name}</span>
            {current.id === org.id && <Check className="h-4 w-4 shrink-0" />}
          </button>
        ))}
        <div className="border-t border-border mt-1 pt-1">
          <button className="flex items-center gap-2 w-full px-3 py-2 rounded-md text-sm text-muted-foreground hover:bg-muted transition-colors">
            <Plus className="h-4 w-4" />
            Create Organization
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
