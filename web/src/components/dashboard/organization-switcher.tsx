"use client";

import * as React from "react";
import { Check, ChevronsUpDown, PlusCircle, Building2 } from "lucide-react";
import { cn } from "~/lib/utils";
import { Button } from "~/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "~/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "~/components/ui/popover";
import { useDashboardShell } from "~/app/[locale]/dashboard/_components/dashboard-shell-context";
import { CreateOrganizationModal } from "./create-organization-modal";

export function OrganizationSwitcher() {
  const [open, setOpen] = React.useState(false);
  const [createModalOpen, setCreateModalOpen] = React.useState(false);
  const { organizations, currentOrganization, setOrganization, refreshOrganizations } = useDashboardShell();

  return (
    <>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            aria-label="Select an organization"
            className={cn(
              "bg-background/50 border-muted-foreground/20 w-[200px] justify-between backdrop-blur-sm",
              "[data-density=compact]:w-[40px] [data-density=compact]:justify-center [data-density=compact]:px-0"
            )}
          >
            <div className="flex items-center gap-2 overflow-hidden">
              <div className="bg-primary/10 text-primary flex h-6 w-6 shrink-0 items-center justify-center rounded-md">
                <Building2 className="h-4 w-4" />
              </div>
              <span className="truncate font-medium [data-density=compact]:hidden">
                {currentOrganization?.name || "Select Org..."}
              </span>
            </div>
            <ChevronsUpDown className="ml-auto h-4 w-4 shrink-0 opacity-50 [data-density=compact]:hidden" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[200px] p-0" align="start">
          <Command>
            <CommandList>
              <CommandInput placeholder="Search organization..." />
              <CommandEmpty>No organization found.</CommandEmpty>
              <CommandGroup heading="Organizations">
                {organizations.map((org) => (
                  <CommandItem
                    key={org.id}
                    onSelect={() => {
                      setOrganization(org);
                      setOpen(false);
                    }}
                    className="cursor-pointer text-sm"
                  >
                    <Building2 className="text-muted-foreground mr-2 h-4 w-4" />
                    {org.name}
                    <Check
                      className={cn(
                        "ml-auto h-4 w-4",
                        currentOrganization?.id === org.id
                          ? "opacity-100"
                          : "opacity-0"
                      )}
                    />
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
            <CommandSeparator />
            <CommandList>
              <CommandGroup>
                <CommandItem
                  onSelect={() => {
                    setOpen(false);
                    setCreateModalOpen(true);
                  }}
                  className="cursor-pointer"
                >
                  <PlusCircle className="mr-2 h-4 w-4" />
                  Create Organization
                </CommandItem>
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      <CreateOrganizationModal
        open={createModalOpen}
        onOpenChange={setCreateModalOpen}
        onSuccess={() => refreshOrganizations()}
      />
    </>
  );
}
