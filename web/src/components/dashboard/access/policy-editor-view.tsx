"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Icon } from "~/components/shared/icon";
import { useDensity } from "~/components/shared/providers/density-provider";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { ScrollArea } from "~/components/ui/scroll-area";
import { Skeleton } from "~/components/ui/skeleton";
import { accessApi, type RoleNode } from "~/lib/api/access";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "~/lib/utils";

interface PolicyEditorViewProps {
  onRoleClick?: (roleId: string, roleName: string) => void;
}

function CRUDLabels({ permissions }: { permissions: string[][] }) {
  const methodSet = new Set(permissions.map((p) => (p[3] ?? "").toUpperCase()));

  const crud = [
    { key: "POST", label: "C", color: "text-emerald-500", active: methodSet.has("POST") },
    { key: "GET", label: "R", color: "text-blue-500", active: methodSet.has("GET") },
    {
      key: "PUT",
      label: "U",
      color: "text-amber-500",
      active: methodSet.has("PUT") || methodSet.has("PATCH"),
    },
    { key: "DELETE", label: "D", color: "text-red-500", active: methodSet.has("DELETE") },
  ];

  return (
    <div className="flex gap-1">
      {crud.map((c) => (
        <span
          key={c.key}
          className={cn(
            "flex h-5 w-5 items-center justify-center rounded text-[10px] font-bold transition-colors",
            c.active 
              ? cn("bg-muted-foreground/10", c.color)
              : "bg-muted/30 text-muted-foreground/20"
          )}
        >
          {c.label}
        </span>
      ))}
    </div>
  );
}

function groupPermissionsByResource(
  permissions: string[][]
): Map<string, string[][]> {
  const map = new Map<string, string[][]>();
  for (const perm of permissions) {
    if (perm.length < 4) continue;
    const path = perm[2] ?? "";
    const parts = path
      .replace(/^\/api\/v\d+\//, "/")
      .split("/")
      .filter(Boolean);
    const resource = "/" + (parts[0] ?? "unknown");

    if (!map.has(resource)) map.set(resource, []);
    map.get(resource)!.push(perm);
  }
  return map;
}

function RoleTreeNode({
  node,
  depth,
  expandedNodes,
  toggleExpand,
  onRoleClick,
  isLast = false,
}: {
  node: RoleNode;
  depth: number;
  expandedNodes: Set<string>;
  toggleExpand: (id: string) => void;
  onRoleClick?: (id: string, name: string) => void;
  isLast?: boolean;
}) {
  const { density } = useDensity();
  const isCompact = density === "compact";
  const isExpanded = expandedNodes.has(node.name);
  const hasChildren = (node.children?.length ?? 0) > 0;
  
  const ownResources = groupPermissionsByResource(node.own_permissions);
  const inheritedResources = groupPermissionsByResource(node.inherited_permissions);
  const allResourceKeys = Array.from(new Set([...ownResources.keys(), ...inheritedResources.keys()]));

  const cleanName = node.name.replace("role:", "");

  return (
    <div className="relative">
      {/* Visual Lines */}
      {depth > 0 && (
        <div 
          className="absolute -left-4 top-0 h-5 w-4 border-b-2 border-l-2 border-muted-foreground/20 rounded-bl-lg" 
          style={{ height: '1.25rem' }}
        />
      )}

      <div
        className={cn(
          "group relative flex items-center gap-2 rounded-lg transition-all duration-200 hover:bg-accent/50",
          isCompact ? "mb-0.5 px-2 py-1 text-xs" : "mb-1 px-3 py-2"
        )}
      >
        <button
          type="button"
          onClick={() => toggleExpand(node.name)}
          className="flex items-center gap-2 text-left outline-none"
        >
          <div className={cn(
            "flex h-6 w-6 items-center justify-center rounded-md transition-colors",
            isExpanded ? "bg-primary/10 text-primary" : "text-muted-foreground group-hover:text-foreground"
          )}>
            <Icon
              name={isExpanded ? "FolderOpen" : "Folder"}
              size={isCompact ? "sm" : "md"}
            />
          </div>
          <span className="font-semibold tracking-tight">{cleanName}</span>
          {hasChildren && (
            <Icon
              name="ChevronRight"
              className={cn(
                "h-3 w-3 text-muted-foreground/50 transition-transform duration-200",
                isExpanded && "rotate-90"
              )}
            />
          )}
        </button>

        {!node.parent_id && (
          <Badge variant="secondary" className="h-4 px-1.5 text-[9px] font-bold uppercase tracking-wider">
            Root
          </Badge>
        )}

        <div className="ml-auto flex items-center gap-3 opacity-0 transition-opacity group-hover:opacity-100">
          <span className="text-muted-foreground/40 text-[10px] font-mono">ID: {node.id.slice(0, 8)}</span>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 px-2 text-xs"
            onClick={(e) => {
              e.stopPropagation();
              onRoleClick?.(node.id, node.name);
            }}
          >
            Edit
          </Button>
        </div>
      </div>

      <AnimatePresence>
        {isExpanded && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            className="ml-8 overflow-hidden"
          >
            {/* Permission List for this role */}
            <div className="mb-2 space-y-1 border-l-2 border-muted-foreground/10 py-1 pl-4">
              {allResourceKeys.length > 0 ? (
                allResourceKeys.map((resource) => {
                  const ownPerms = ownResources.get(resource) ?? [];
                  const inhPerms = inheritedResources.get(resource) ?? [];
                  const isOwn = ownPerms.length > 0;
                  const isInherited = inhPerms.length > 0;
                  const effectivePerms = isOwn ? ownPerms : inhPerms;

                  return (
                    <div
                      key={resource}
                      className="flex items-center justify-between rounded-md px-2 py-1 hover:bg-muted/30"
                    >
                      <div className="flex items-center gap-2">
                        <div className="h-1 w-1 rounded-full bg-muted-foreground/30" />
                        <span className="font-mono text-[11px] text-muted-foreground">{resource}</span>
                        {isOwn && isInherited && (
                          <span className="bg-amber-500/10 text-amber-600 rounded px-1 text-[9px] font-bold uppercase">Override</span>
                        )}
                        {!isOwn && isInherited && (
                          <span className="text-muted-foreground/40 text-[9px] italic">(inherited)</span>
                        )}
                      </div>
                      <CRUDLabels permissions={effectivePerms} />
                    </div>
                  );
                })
              ) : (
                <div className="px-2 py-1 text-[10px] italic text-muted-foreground/50">
                  No explicit permissions
                </div>
              )}
            </div>

            {/* Child Roles */}
            {node.children?.map((child, idx) => (
              <RoleTreeNode
                key={child.name}
                node={child}
                depth={depth + 1}
                expandedNodes={expandedNodes}
                toggleExpand={toggleExpand}
                onRoleClick={onRoleClick}
                isLast={idx === (node.children?.length ?? 0) - 1}
              />
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

export function PolicyEditorView({ onRoleClick }: PolicyEditorViewProps) {
  const [treeData, setTreeData] = useState<RoleNode[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    try {
      const resp = await accessApi.getInheritanceTree();
      if (resp.data?.roles) {
        setTreeData(resp.data.roles);
        // Default expand first level
        const roots = resp.data.roles.map(r => r.name);
        setExpandedNodes(new Set(roots));
      }
    } catch {
      toast.error("Failed to load inheritance tree");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const toggleExpand = (name: string) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (next.has(name)) next.delete(name);
      else next.add(name);
      return next;
    });
  };

  const expandAll = () => {
    const allNames = new Set<string>();
    const collect = (nodes: RoleNode[]) => {
      for (const n of nodes) {
        allNames.add(n.name);
        if (n.children) collect(n.children);
      }
    };
    collect(treeData);
    setExpandedNodes(allNames);
  };

  const collapseAll = () => setExpandedNodes(new Set());

  if (isLoading) {
    return (
      <div className="space-y-4 py-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full rounded-xl" />
        ))}
      </div>
    );
  }

  if (treeData.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-xl border border-dashed bg-muted/5 py-24 text-center">
        <div className="bg-muted/10 mb-4 flex h-16 w-16 items-center justify-center rounded-full ring-8 ring-muted/5">
          <Icon name="GitBranch" className="text-muted-foreground/30 h-8 w-8" />
        </div>
        <h3 className="text-lg font-semibold">No Role Hierarchy</h3>
        <p className="text-muted-foreground mt-1 max-w-xs text-sm">
          Start by creating roles and defining parent-child relationships to see the tree.
        </p>
      </div>
    );
  }

  return (
    <div className="bg-card rounded-xl border shadow-sm">
      <div className="flex items-center justify-between border-b px-5 py-4">
        <div className="flex items-center gap-3">
          <div className="bg-primary/10 flex h-8 w-8 items-center justify-center rounded-lg">
            <Icon name="GitBranch" className="text-primary h-5 w-5" />
          </div>
          <div>
            <h3 className="text-sm font-bold">Inheritance Explorer</h3>
            <p className="text-muted-foreground text-[10px] uppercase font-bold tracking-widest">RBAC Policy Tree</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" className="h-8 px-3 text-xs font-semibold" onClick={expandAll}>
            Expand All
          </Button>
          <Button variant="ghost" size="sm" className="h-8 px-2 text-xs" onClick={collapseAll}>
            Collapse All
          </Button>
        </div>
      </div>

      <ScrollArea className="h-[600px]">
        <div className="p-6">
          <div className="space-y-2">
            {treeData.map((node) => (
              <RoleTreeNode
                key={node.name}
                node={node}
                depth={0}
                expandedNodes={expandedNodes}
                toggleExpand={toggleExpand}
                onRoleClick={onRoleClick}
              />
            ))}
          </div>
        </div>
      </ScrollArea>

      <div className="bg-muted/30 flex items-center justify-between border-t px-5 py-3">
        <div className="flex items-center gap-6 text-[10px] font-bold tracking-widest text-muted-foreground uppercase">
          <div className="flex items-center gap-1.5">
            <div className="h-2 w-2 rounded-sm bg-emerald-500/20 border border-emerald-500/50" />
            <span>Create</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="h-2 w-2 rounded-sm bg-blue-500/20 border border-blue-500/50" />
            <span>Read</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="h-2 w-2 rounded-sm bg-amber-500/20 border border-amber-500/50" />
            <span>Update</span>
          </div>
          <div className="flex items-center gap-1.5">
            <div className="h-2 w-2 rounded-sm bg-red-500/20 border border-red-500/50" />
            <span>Delete</span>
          </div>
        </div>
        <div className="text-muted-foreground flex items-center gap-2 text-[10px] italic">
          <Icon name="Info" className="h-3 w-3" />
          <span>Permissions are calculated top-down</span>
        </div>
      </div>
    </div>
  );
}
