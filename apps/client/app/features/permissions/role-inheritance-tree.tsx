import { cn } from "@casbin/ui";
import { NexusCard } from "@casbin/ui";
import { Badge } from "@casbin/ui";
import { ChevronRight, Shield } from "lucide-react";
import { useState } from "react";

export interface RoleNode {
  id: string;
  name: string;
  description?: string;
  effective_permissions?: string[][];
  children?: RoleNode[];
}

interface RoleInheritanceTreeProps {
  tree: RoleNode[];
  onSelect?: (role: RoleNode) => void;
}

function TreeNode({
  node,
  depth,
  onSelect,
}: {
  node: RoleNode;
  depth: number;
  onSelect?: (r: RoleNode) => void;
}) {
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
          "hover:bg-muted/50 group flex w-full items-center gap-2 rounded-md px-3 py-2.5 text-sm transition-colors",
          "text-foreground",
        )}
        style={{ paddingLeft: `${depth * 20 + 12}px` }}
      >
        {hasChildren ? (
          <ChevronRight
            className={cn(
              "text-muted-foreground h-4 w-4 shrink-0 transition-transform duration-200",
              expanded && "rotate-90",
            )}
          />
        ) : (
          <span className="w-4" />
        )}
        <Shield className="text-primary h-4 w-4 shrink-0" />
        <span className="font-medium">{node.name}</span>
        <Badge variant="secondary" className="ml-auto text-[10px]">
          {(node.effective_permissions || []).length} perms
        </Badge>
      </button>
      {expanded && hasChildren && (
        <div className="border-border ml-6 border-l">
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
      <div className="border-border border-b p-4">
        <h3 className="text-foreground text-sm font-semibold">Role Inheritance</h3>
        <p className="text-muted-foreground mt-0.5 text-xs">Visual hierarchy of role permissions</p>
      </div>
      <div className="py-2">
        {tree.map((node) => (
          <TreeNode key={node.id} node={node} depth={0} onSelect={onSelect} />
        ))}
      </div>
    </NexusCard>
  );
}
