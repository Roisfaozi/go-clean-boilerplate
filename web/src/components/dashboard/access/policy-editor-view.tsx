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

interface PolicyEditorViewProps {
  onRoleClick?: (roleId: string, roleName: string) => void;
}

function CRUDLabel({ permissions }: { permissions: string[][] }) {
  const methodSet = new Set(permissions.map((p) => (p[3] ?? "").toUpperCase()));

  const crud = [
    { key: "POST", label: "C", active: methodSet.has("POST") },
    { key: "GET", label: "R", active: methodSet.has("GET") },
    {
      key: "PUT",
      label: "U",
      active: methodSet.has("PUT") || methodSet.has("PATCH"),
    },
    { key: "DELETE", label: "D", active: methodSet.has("DELETE") },
  ];

  return (
    <span className="font-mono text-xs">
      {crud.map((c) => (
        <span
          key={c.key}
          className={
            c.active
              ? "text-foreground font-semibold"
              : "text-muted-foreground/40"
          }
        >
          {c.active ? c.label : "-"}
        </span>
      ))}
    </span>
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
}: {
  node: RoleNode;
  depth: number;
  expandedNodes: Set<string>;
  toggleExpand: (id: string) => void;
  onRoleClick?: (id: string, name: string) => void;
}) {
  const { density } = useDensity();
  const isCompact = density === "compact";
  const indentSize = isCompact ? 16 : 20;

  const isExpanded = expandedNodes.has(node.name);
  const hasChildren = (node.children?.length ?? 0) > 0;
  const ownResources = groupPermissionsByResource(node.own_permissions);
  const inheritedResources = groupPermissionsByResource(
    node.inherited_permissions
  );

  const allResourceKeys = new Set([
    ...ownResources.keys(),
    ...inheritedResources.keys(),
  ]);

  const cleanName = node.name.replace("role:", "");

  return (
    <div className="select-none">
      <div
        className={`group hover:bg-muted/50 flex items-center gap-1 rounded-md transition-colors ${isCompact ? "px-1.5 py-1 text-xs" : "px-2 py-1.5"}`}
        style={{ paddingLeft: `${depth * indentSize + (isCompact ? 4 : 8)}px` }}
      >
        {depth > 0 && (
          <div className="text-muted-foreground/30 mr-1 flex items-center">
            {Array.from({ length: depth }).map((_, i) => (
              <span
                key={i}
                className="border-muted-foreground/20 mr-3 inline-block h-full border-l"
              />
            ))}
            <span className="mr-1">├──</span>
          </div>
        )}

        <button
          type="button"
          onClick={() => toggleExpand(node.name)}
          className="flex items-center gap-1.5 text-left"
        >
          <Icon
            name={isExpanded ? "FolderOpen" : "Folder"}
            className={`text-primary/70 flex-shrink-0 ${isCompact ? "h-3.5 w-3.5" : "h-4 w-4"}`}
          />
          <span className="font-medium">{cleanName}</span>
          <Icon
            name={isExpanded ? "ChevronDown" : "ChevronRight"}
            className={`text-muted-foreground ${isCompact ? "h-2.5 w-2.5" : "h-3 w-3"}`}
          />
        </button>

        {!node.parent_id && (
          <Badge
            variant="outline"
            className={`ml-2 px-1.5 py-0 ${isCompact ? "text-[8px]" : "text-[9px]"}`}
          >
            Root
          </Badge>
        )}

        {node.parent_id && (
          <span
            className={`text-muted-foreground/50 ml-2 ${isCompact ? "text-[9px]" : "text-[10px]"}`}
          >
            ← inherits from {node.parent_id.replace("role:", "")}
          </span>
        )}

        <div className="ml-auto flex items-center gap-2 opacity-0 transition-opacity group-hover:opacity-100">
          <Button
            variant="ghost"
            size="sm"
            className={`px-2 ${isCompact ? "h-5 text-[10px]" : "h-6 text-xs"}`}
            onClick={(e) => {
              e.stopPropagation();
              onRoleClick?.(node.id, node.name);
            }}
          >
            Edit
          </Button>
        </div>
      </div>

      {isExpanded && (
        <div className="ml-1">
          {allResourceKeys.size > 0 ? (
            Array.from(allResourceKeys).map((resource) => {
              const ownPerms = ownResources.get(resource) ?? [];
              const inhPerms = inheritedResources.get(resource) ?? [];
              const isOwn = ownPerms.length > 0;
              const isInherited = inhPerms.length > 0;

              const effectivePerms = isOwn ? ownPerms : inhPerms;

              let sourceLabel = "";
              if (isOwn && isInherited) sourceLabel = "override";
              else if (isOwn) sourceLabel = "own permission";
              else if (isInherited) sourceLabel = "inherited";

              return (
                <div
                  key={resource}
                  className={`flex items-center gap-2 ${isCompact ? "py-0.5" : "py-1"} text-xs`}
                  style={{
                    paddingLeft: `${(depth + 1) * indentSize + (isCompact ? 16 : 24)}px`,
                  }}
                >
                  <Icon
                    name="Lock"
                    className={`flex-shrink-0 text-amber-500/70 ${isCompact ? "h-2.5 w-2.5" : "h-3 w-3"}`}
                  />
                  <code className="text-muted-foreground">{resource}:</code>
                  <CRUDLabel permissions={effectivePerms} />
                  <span className="text-muted-foreground/50 text-[10px]">
                    ({sourceLabel})
                  </span>
                </div>
              );
            })
          ) : (
            <div
              className={`text-muted-foreground/50 ${isCompact ? "py-0.5" : "py-1"} text-xs italic`}
              style={{
                paddingLeft: `${(depth + 1) * indentSize + (isCompact ? 16 : 24)}px`,
              }}
            >
              No permissions defined
            </div>
          )}

          {node.children?.map((child) => (
            <RoleTreeNode
              key={child.name}
              node={child}
              depth={depth + 1}
              expandedNodes={expandedNodes}
              toggleExpand={toggleExpand}
              onRoleClick={onRoleClick}
            />
          ))}
        </div>
      )}
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
        const allNames = new Set<string>();
        const collectNames = (nodes: RoleNode[]) => {
          for (const n of nodes) {
            allNames.add(n.name);
            if (n.children) collectNames(n.children);
          }
        };
        collectNames(resp.data.roles);
        setExpandedNodes(allNames);
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
      <div className="space-y-3 py-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-8 w-full" />
        ))}
      </div>
    );
  }

  if (treeData.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-lg border-2 border-dashed py-16">
        <Icon
          name="GitBranch"
          className="text-muted-foreground/30 mb-3 h-10 w-10"
        />
        <p className="text-muted-foreground text-sm">
          No role hierarchy found. Add role inheritance relationships first.
        </p>
      </div>
    );
  }

  return (
    <div className="rounded-lg border">
      <div className="flex items-center justify-between border-b px-4 py-3">
        <div className="flex items-center gap-2">
          <Icon name="GitBranch" className="text-primary h-4 w-4" />
          <h3 className="text-sm font-semibold">Role Inheritance Tree</h3>
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="h-7 text-xs"
            onClick={expandAll}
          >
            Expand All
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-7 text-xs"
            onClick={collapseAll}
          >
            <Icon name="Minus" className="h-3 w-3" />
          </Button>
        </div>
      </div>

      <ScrollArea className="max-h-[600px]">
        <div className="space-y-1 p-4">
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
      </ScrollArea>

      <div className="text-muted-foreground flex flex-wrap items-center gap-4 border-t px-4 py-3 text-[11px]">
        <div className="flex items-center gap-1">
          <Icon name="Folder" className="text-primary/70 h-3 w-3" />
          <span>Role node</span>
        </div>
        <div className="flex items-center gap-1">
          <Icon name="Lock" className="h-3 w-3 text-amber-500/70" />
          <span>Permission</span>
        </div>
        <div>C=Create R=Read U=Update D=Delete -=Denied</div>
      </div>
    </div>
  );
}
