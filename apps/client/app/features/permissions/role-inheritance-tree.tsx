import { cn, NexusBadge, NexusCard } from "@casbin/ui";
import { ChevronRight, Shield } from "lucide-react";
import { useState } from "react";

export interface RoleNode {
	id: string;
	name: string;
	description?: string;
	parents?: string[];
	own_permissions?: string[][];
	inherited_permissions?: string[][];
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
	activeId,
}: {
	node: RoleNode;
	depth: number;
	onSelect?: (r: RoleNode) => void;
	activeId?: string;
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
					activeId === node.id
						? "bg-primary/10 text-primary"
						: "text-foreground",
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
				<NexusBadge variant="neutral" className="ml-auto text-[10px]">
					{(node.effective_permissions || []).length} perms
				</NexusBadge>
			</button>
			{expanded && hasChildren && (
				<div className="border-border ml-6 border-l">
					{node.children!.map((child) => (
						<TreeNode
							key={child.id}
							node={child}
							depth={depth + 1}
							onSelect={onSelect}
							activeId={activeId}
						/>
					))}
				</div>
			)}
		</div>
	);
}

export function RoleInheritanceTree({
	tree,
	onSelect,
	activeId,
}: RoleInheritanceTreeProps & { activeId?: string }) {
	return (
		<NexusCard className="shadow-premium h-full overflow-hidden border-none">
			<div className="border-border bg-muted/20 border-b p-4">
				<h3 className="text-foreground text-sm font-semibold">
					Security Hierarchy
				</h3>
				<p className="text-muted-foreground mt-0.5 text-xs">
					Visual tree of role permissions and inheritance
				</p>
			</div>
			<div className="max-h-[600px] overflow-y-auto py-2">
				{tree.length === 0 ? (
					<div className="flex flex-col items-center justify-center py-10 text-center">
						<Shield className="text-muted/30 mb-2 h-10 w-10" />
						<p className="text-muted-foreground text-xs italic">
							No roles defined
						</p>
					</div>
				) : (
					tree.map((node) => (
						<TreeNode
							key={node.id}
							node={node}
							depth={0}
							onSelect={onSelect}
							activeId={activeId}
						/>
					))
				)}
			</div>
		</NexusCard>
	);
}
