import { cn } from "@casbin/ui";
import { ChevronRight, ChevronDown } from "lucide-react";
import { useState } from "react";

/* ── TreeView ── */
export interface TreeNode {
	id: string;
	label: string;
	icon?: React.ReactNode;
	children?: TreeNode[];
}

interface TreeViewProps {
	nodes: TreeNode[];
	onSelect?: (node: TreeNode) => void;
	className?: string;
}

export function TreeView({ nodes, onSelect, className }: TreeViewProps) {
	return (
		<div className={cn("space-y-0.5", className)}>
			{nodes.map((node) => (
				<TreeNode key={node.id} node={node} onSelect={onSelect} depth={0} />
			))}
		</div>
	);
}

function TreeNode({
	node,
	onSelect,
	depth,
}: {
	node: TreeNode;
	onSelect?: (n: TreeNode) => void;
	depth: number;
}) {
	const [open, setOpen] = useState(false);
	const hasChildren = node.children && node.children.length > 0;

	return (
		<div>
			<button
				onClick={() => {
					if (hasChildren) setOpen(!open);
					onSelect?.(node);
				}}
				className="text-body hover:bg-surface-hover flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors"
				style={{ paddingLeft: `${depth * 20 + 8}px` }}
			>
				{hasChildren ? (
					open ? (
						<ChevronDown className="text-muted-foreground h-4 w-4 shrink-0" />
					) : (
						<ChevronRight className="text-muted-foreground h-4 w-4 shrink-0" />
					)
				) : (
					<span className="w-4" />
				)}
				{node.icon && <span className="shrink-0">{node.icon}</span>}
				<span className="truncate">{node.label}</span>
			</button>
			{hasChildren && open && (
				<div>
					{node.children!.map((child) => (
						<TreeNode
							key={child.id}
							node={child}
							onSelect={onSelect}
							depth={depth + 1}
						/>
					))}
				</div>
			)}
		</div>
	);
}

/* ── KeyValueList ── */
interface KeyValueListProps {
	items: { key: string; value: React.ReactNode }[];
	striped?: boolean;
	className?: string;
}

export function KeyValueList({ items, striped, className }: KeyValueListProps) {
	return (
		<dl
			className={cn(
				"divide-border border-border divide-y overflow-hidden rounded-lg border",
				className,
			)}
		>
			{items.map((item, i) => (
				<div
					key={item.key}
					className={cn(
						"flex items-start justify-between gap-4 px-4 py-3",
						striped && i % 2 === 0 && "bg-surface",
					)}
				>
					<dt className="text-small text-muted-foreground shrink-0 font-medium">
						{item.key}
					</dt>
					<dd className="text-body text-foreground text-right">{item.value}</dd>
				</div>
			))}
		</dl>
	);
}

/* ── CodeBlock ── */
interface CodeBlockProps {
	code: string;
	language?: string;
	className?: string;
}

export function CodeBlock({ code, language, className }: CodeBlockProps) {
	return (
		<div
			className={cn(
				"border-border bg-surface relative overflow-hidden rounded-lg border",
				className,
			)}
		>
			{language && (
				<div className="border-border flex items-center justify-between border-b px-4 py-2">
					<span className="text-caption text-muted-foreground font-mono">
						{language}
					</span>
				</div>
			)}
			<pre className="overflow-x-auto p-4">
				<code className="text-small text-foreground font-mono whitespace-pre">
					{code}
				</code>
			</pre>
		</div>
	);
}

/* ── Timeline ── */
export interface TimelineItem {
	id: string;
	title: string;
	description?: string;
	time: string;
	icon?: React.ReactNode;
	variant?: "default" | "success" | "danger" | "warning" | "info";
}

interface TimelineProps {
	items: TimelineItem[];
	className?: string;
}

const timelineColors = {
	default: "bg-muted-foreground",
	success: "bg-success",
	danger: "bg-danger",
	warning: "bg-warning",
	info: "bg-info",
};

export function Timeline({ items, className }: TimelineProps) {
	return (
		<div className={cn("space-y-0", className)}>
			{items.map((item, i) => (
				<div key={item.id} className="flex gap-4">
					<div className="flex flex-col items-center">
						<div
							className={cn(
								"flex h-8 w-8 shrink-0 items-center justify-center rounded-full",
								timelineColors[item.variant || "default"],
							)}
						>
							{item.icon || (
								<div className="bg-background h-2 w-2 rounded-full" />
							)}
						</div>
						{i < items.length - 1 && <div className="bg-border w-px flex-1" />}
					</div>
					<div className="pt-1 pb-6">
						<p className="text-body text-foreground font-medium">
							{item.title}
						</p>
						{item.description && (
							<p className="text-small text-muted-foreground mt-0.5">
								{item.description}
							</p>
						)}
						<p className="text-caption text-muted-foreground mt-1">
							{item.time}
						</p>
					</div>
				</div>
			))}
		</div>
	);
}

/* ── ActivityFeed ── */
export interface ActivityItem {
	id: string;
	user: { name: string; avatar?: string };
	action: string;
	target?: string;
	time: string;
}

interface ActivityFeedProps {
	items: ActivityItem[];
	className?: string;
}

export function ActivityFeed({ items, className }: ActivityFeedProps) {
	return (
		<div className={cn("divide-border divide-y", className)}>
			{items.map((item) => (
				<div key={item.id} className="flex items-start gap-3 py-3">
					<div className="bg-primary/10 text-primary text-caption flex h-8 w-8 shrink-0 items-center justify-center rounded-full font-bold">
						{item.user.avatar ? (
							<img
								src={item.user.avatar}
								alt={item.user.name}
								className="h-8 w-8 rounded-full object-cover"
							/>
						) : (
							item.user.name.charAt(0).toUpperCase()
						)}
					</div>
					<div className="min-w-0 flex-1">
						<p className="text-body">
							<span className="text-foreground font-semibold">
								{item.user.name}
							</span>{" "}
							<span className="text-muted-foreground">{item.action}</span>{" "}
							{item.target && (
								<span className="text-foreground font-medium">
									{item.target}
								</span>
							)}
						</p>
						<p className="text-caption text-muted-foreground">{item.time}</p>
					</div>
				</div>
			))}
		</div>
	);
}
