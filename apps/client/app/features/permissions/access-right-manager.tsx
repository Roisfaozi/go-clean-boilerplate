import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { NexusCard } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { useToast } from "@casbin/ui";
import { Search } from "lucide-react";
import { accessApi } from "@/lib/api/access";

export function AccessRightManager() {
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const [search, setSearch] = useState("");
  const [selectedRightId, setSelectedRightId] = useState<string | null>(null);

  const { data: rightsRes, isLoading: isLoadingRights } = useQuery({
    queryKey: ["access-rights"],
    queryFn: () => accessApi.listRights({ limit: 100 }),
  });

  const { data: endpointsRes } = useQuery({
    queryKey: ["endpoints"],
    queryFn: () => accessApi.listEndpoints({ limit: 500 }),
  });

  const _linkMutation = useMutation({
    mutationFn: accessApi.link,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["access-rights"] });
      toast({ title: "Endpoint Linked" });
    },
  });

  const _unlinkMutation = useMutation({
    mutationFn: accessApi.unlink,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["access-rights"] });
      toast({ title: "Endpoint Unlinked" });
    },
  });

  const rights = rightsRes?.data || [];
  const _allEndpoints = endpointsRes?.data || [];

  const filtered = rights.filter((r) =>
    r.name.toLowerCase().includes(search.toLowerCase()),
  );

  const selected = rights.find((r) => r.id === selectedRightId);

  // Note: We need a way to know which endpoints are currently linked to which access right.
  // The current backend might return this in the access right object or we might need another call.
  // For now, let's assume linked endpoints are not directly in AccessRight and we might need to fetch them.
  // Wait! Let's check the backend model again.
  // It didn't have endpoints in AccessRight.

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
      {/* Rights list */}
      <NexusCard className="lg:col-span-1">
        <div className="border-border border-b p-4">
          <h3 className="text-foreground text-sm font-semibold">
            Access Rights
          </h3>
          <div className="relative mt-2">
            <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
            <NexusInput
              placeholder="Search rights..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-8 pl-9 text-sm"
            />
          </div>
        </div>
        <div className="divide-border max-h-[400px] divide-y overflow-y-auto">
          {isLoadingRights ? (
            <div className="text-muted-foreground p-4 text-center text-xs">
              Loading...
            </div>
          ) : (
            filtered.map((right) => (
              <button
                key={right.id}
                onClick={() => setSelectedRightId(right.id)}
                className={`hover:bg-muted/50 w-full px-4 py-3 text-left text-sm transition-colors ${
                  selectedRightId === right.id
                    ? "bg-primary/5 border-l-primary border-l-2"
                    : ""
                }`}
              >
                <div className="text-foreground font-medium">{right.name}</div>
                <div className="text-muted-foreground mt-0.5 text-xs">
                  {right.resource}:{right.action}
                </div>
              </button>
            ))
          )}
        </div>
      </NexusCard>

      {/* Linked endpoints */}
      <NexusCard className="lg:col-span-2">
        {selected ? (
          <div className="text-muted-foreground p-4 text-center text-sm">
            Access Right management is currently being migrated to use granular
            Endpoint mapping.
            <br />
            Please use the Permission Matrix for direct role-to-resource
            mapping.
          </div>
        ) : (
          <div className="text-muted-foreground flex h-64 items-center justify-center text-sm">
            Select an access right to manage
          </div>
        )}
      </NexusCard>
    </div>
  );
}
