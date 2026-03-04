"use client";

import {
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  Search,
  Shield,
} from "lucide-react";
import { useEffect, useState, useTransition } from "react";
import { toast } from "sonner";
import { Badge } from "~/components/ui/badge";
import { Input } from "~/components/ui/input";
import { ScrollArea } from "~/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "~/components/ui/sheet";
import { Skeleton } from "~/components/ui/skeleton";
import { Switch } from "~/components/ui/switch";
import { accessApi, RoleAccessRightStatus } from "~/lib/api/access";
import { cn } from "~/lib/utils";

interface RolePermissionSheetProps {
  roleName: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  domain?: string;
}

export function RolePermissionSheet({
  roleName,
  open,
  onOpenChange,
  domain = "global",
}: RolePermissionSheetProps) {
  const [accessRights, setAccessRights] = useState<RoleAccessRightStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});
  const [toggling, setToggling] = useState<Record<string, boolean>>({});
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    if (!open || !roleName) return;
    setSearch("");
    setExpanded({});
    loadAccessRights();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, roleName]);

  const loadAccessRights = async () => {
    if (!roleName) return;
    setLoading(true);
    try {
      const res = await accessApi.getRoleAccessRights(roleName, domain);
      setAccessRights(res.data ?? []);
    } catch {
      toast.error("Failed to load access rights");
    } finally {
      setLoading(false);
    }
  };

  const handleToggle = async (ar: RoleAccessRightStatus, newValue: boolean) => {
    if (!roleName || toggling[ar.id]) return;
    setToggling((prev) => ({ ...prev, [ar.id]: true }));

    try {
      if (newValue) {
        await accessApi.assignAccessRight(roleName, ar.id, domain);
        toast.success(`Assigned "${ar.name}" to ${roleName}`);
      } else {
        await accessApi.revokeAccessRight(roleName, ar.id, domain);
        toast.success(`Revoked "${ar.name}" from ${roleName}`);
      }
      // Optimistically update state
      setAccessRights((prev) =>
        prev.map((item) =>
          item.id === ar.id
            ? { ...item, is_assigned: newValue, is_partial: false }
            : item
        )
      );
    } catch {
      toast.error(`Failed to ${newValue ? "assign" : "revoke"} "${ar.name}"`);
    } finally {
      setToggling((prev) => ({ ...prev, [ar.id]: false }));
    }
  };

  const toggleExpand = (id: string) => {
    setExpanded((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const filtered = accessRights.filter((ar) =>
    ar.name.toLowerCase().includes(search.toLowerCase())
  );

  const assignedCount = accessRights.filter((ar) => ar.is_assigned).length;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex w-full flex-col gap-0 p-0 sm:max-w-lg">
        {/* Header */}
        <SheetHeader className="border-border/50 border-b px-6 pt-6 pb-4">
          <div className="flex items-center gap-3">
            <div className="bg-primary/10 rounded-lg p-2">
              <Shield className="text-primary h-5 w-5" />
            </div>
            <div>
              <SheetTitle className="text-lg font-semibold">
                Manage Permissions
              </SheetTitle>
              <SheetDescription className="text-muted-foreground mt-0.5 text-sm">
                <span className="bg-muted rounded px-1.5 py-0.5 font-mono text-xs">
                  {roleName}
                </span>
              </SheetDescription>
            </div>
          </div>
        </SheetHeader>

        {/* Search */}
        <div className="border-border/50 border-b px-6 py-3">
          <div className="relative">
            <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
            <Input
              placeholder="Search access rights..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="bg-muted/50 h-9 pl-9"
            />
          </div>
        </div>

        {/* List */}
        <ScrollArea className="flex-1 px-2">
          <div className="space-y-1 py-2">
            {loading ? (
              Array.from({ length: 6 }).map((_, i) => (
                <div
                  key={i}
                  className="flex items-center justify-between px-4 py-3"
                >
                  <div className="flex-1 space-y-1.5">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-20" />
                  </div>
                  <Skeleton className="h-5 w-9 rounded-full" />
                </div>
              ))
            ) : filtered.length === 0 ? (
              <div className="text-muted-foreground flex flex-col items-center justify-center py-12">
                <Shield className="mb-2 h-8 w-8 opacity-30" />
                <p className="text-sm">No access rights found</p>
              </div>
            ) : (
              filtered.map((ar) => (
                <div
                  key={ar.id}
                  className="border-border/40 mx-2 overflow-hidden rounded-lg border"
                >
                  {/* Access Right Row */}
                  <div className="hover:bg-muted/30 flex items-center gap-3 px-4 py-3 transition-colors">
                    {/* Expand toggle */}
                    <button
                      onClick={() => toggleExpand(ar.id)}
                      className="text-muted-foreground hover:text-foreground shrink-0 transition-colors"
                    >
                      {expanded[ar.id] ? (
                        <ChevronDown className="h-4 w-4" />
                      ) : (
                        <ChevronRight className="h-4 w-4" />
                      )}
                    </button>

                    {/* Name + badges */}
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="truncate text-sm font-medium">
                          {ar.name}
                        </span>
                        <Badge variant="secondary" className="shrink-0 text-xs">
                          {ar.endpoints.length} endpoint
                          {ar.endpoints.length !== 1 ? "s" : ""}
                        </Badge>
                        {ar.is_partial && (
                          <Badge
                            variant="outline"
                            className="shrink-0 border-amber-500/50 text-xs text-amber-500"
                          >
                            <AlertTriangle className="mr-1 h-3 w-3" />
                            Partial
                          </Badge>
                        )}
                      </div>
                    </div>

                    {/* Toggle */}
                    <Switch
                      checked={ar.is_assigned}
                      onCheckedChange={(val) => handleToggle(ar, val)}
                      disabled={toggling[ar.id] || isPending}
                      className={cn(
                        "shrink-0 transition-colors",
                        toggling[ar.id] && "opacity-50"
                      )}
                    />
                  </div>

                  {/* Endpoints expanded */}
                  {expanded[ar.id] && ar.endpoints.length > 0 && (
                    <div className="border-border/40 bg-muted/20 space-y-1 border-t px-4 py-2">
                      {ar.endpoints.map((ep, i) => {
                        const [method, ...pathParts] = ep.split(" ");
                        const path = pathParts.join(" ");
                        return (
                          <div
                            key={i}
                            className="flex items-center gap-2 text-xs"
                          >
                            <span
                              className={cn(
                                "w-14 rounded px-1 py-0.5 text-center font-mono font-semibold",
                                method === "GET" &&
                                  "bg-emerald-400/10 text-emerald-400",
                                method === "POST" &&
                                  "bg-blue-400/10 text-blue-400",
                                method === "PUT" &&
                                  "bg-amber-400/10 text-amber-400",
                                method === "PATCH" &&
                                  "bg-orange-400/10 text-orange-400",
                                method === "DELETE" &&
                                  "bg-red-400/10 text-red-400"
                              )}
                            >
                              {method}
                            </span>
                            <span className="text-muted-foreground truncate font-mono">
                              {path}
                            </span>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        </ScrollArea>

        {/* Footer */}
        {!loading && accessRights.length > 0 && (
          <div className="border-border/50 border-t px-6 py-3">
            <p className="text-muted-foreground text-center text-xs">
              <span className="text-foreground font-semibold">
                {assignedCount}
              </span>{" "}
              of{" "}
              <span className="text-foreground font-semibold">
                {accessRights.length}
              </span>{" "}
              access rights assigned
            </p>
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
