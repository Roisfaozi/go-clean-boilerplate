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
import { organizationsApi, Organization } from "~/lib/api/organizations";
import { useOrganizationStore } from "~/stores/use-organization-store";
import { toast } from "sonner";
import { Icon } from "../shared/icon";
import { CreateOrganizationModal } from "./create-organization-modal";

export function OrganizationSwitcher() {
  const [open, setOpen] = React.useState(false);
  const [createModalOpen, setCreateModalOpen] = React.useState(false);
  const [organizations, setOrganizations] = React.useState<Organization[]>([]);
  const { currentOrganization, setCurrentOrganization } = useOrganizationStore();
  const [isLoading, setIsLoading] = React.useState(true);

  const fetchOrgs = React.useCallback(async () => {
    try {
      const resp = await organizationsApi.getMyOrganizations();
      if (resp.data?.organizations) {
        setOrganizations(resp.data.organizations);
        // Auto-select first org if none selected
        if (!currentOrganization && resp.data.organizations.length > 0) {
          setCurrentOrganization(resp.data.organizations[0]);
        }
      }
    } catch (error) {
      console.error("Failed to fetch organizations", error);
    } finally {
      setIsLoading(false);
    }
  }, [currentOrganization, setCurrentOrganization]);

  React.useEffect(() => {
    fetchOrgs();
  }, [fetchOrgs]);

  return (
    <>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            aria-label="Select an organization"
            className={cn("w-[200px] justify-between bg-background/50 backdrop-blur-sm border-muted-foreground/20", 
              "[data-density=compact]:w-[40px] [data-density=compact]:px-0 [data-density=compact]:justify-center")}
          >
            <div className="flex items-center gap-2 overflow-hidden">
              <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
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
                      setCurrentOrganization(org);
                      setOpen(false);
                      toast.success(`Switched to ${org.name}`);
                      // Optional: Refresh page or trigger context update
                      // window.location.reload(); 
                    }}
                    className="text-sm cursor-pointer"
                  >
                    <Building2 className="mr-2 h-4 w-4 text-muted-foreground" />
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
        onSuccess={() => fetchOrgs()}
      />
    </>
  );
}
