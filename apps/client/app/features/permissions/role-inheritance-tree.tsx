import { cn } from "@casbin/ui";
import {  NexusCard  } from "@casbin/ui";
import {  Badge  } from "@casbin/ui";
import { ChevronRight, Shield } from "lucide-react";
import { useState } from "react";

export interface RoleNode {
  id: string;
  name: string;
  permissions: string[];
  children?: RoleNode[];
}

interface RoleInheritanceTreeProps {
  tree: RoleNode[];
  onSelect?: (role: RoleNode) => void;
}

function TreeNode({ node, depth, onSelect }: { node: RoleNode; depth: number; onSelect?: (r: RoleNode) => void }) {
  const [expanded, setExpanded] = useState(depth < 2);
  const hasChildren = node.children && node.children.length > 0;

  return (
    <div>
      <button
        onClick={() => {
          if (hasChildren) setExpanded(!expanded);
          onSelect?.(node);
        }}
        className={cn(
          "flex items-center gap-2 w-full px-3 py-2.5 rounded-md text-sm transition-colors hover:bg-muted/50 group",
          "text-foreground"
        )}
        style={{ paddingLeft: `${depth * 20 + 12}px` }}
      >
        {hasChildren ? (
          <ChevronRight
            className={cn("h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200", expanded && "rotate-90")}
          />
        ) : (
          <span className="w-4" />
        )}
        <Shield className="h-4 w-4 shrink-0 text-primary" />
        <span className="font-medium">{node.name}</span>
        <Badge variant="secondary" className="ml-auto text-[10px]">
          {node.permissions.length} perms
        </Badge>
      </button>
      {expanded && hasChildren && (
        <div className="border-l border-border ml-6">
          {node.children!.map((child) => (
            <TreeNode key={child.id} node={child} depth={depth + 1} onSelect={onSelect} />
          ))}
        </div>
      )}
    </div>
  );
}

export function RoleInheritanceTree({ tree, onSelect }: RoleInheritanceTreeProps) {
  return (
    <NexusCard>
      <div className="p-4 border-b border-border">
        <h3 className="text-sm font-semibold text-foreground">Role Inheritance</h3>
        <p className="text-xs text-muted-foreground mt-0.5">Visual hierarchy of role permissions</p>
      </div>
      <div className="py-2">
        {tree.map((node) => (
          <TreeNode key={node.id} node={node} depth={0} onSelect={onSelect} />
        ))}
      </div>
    </NexusCard>
  );
}
