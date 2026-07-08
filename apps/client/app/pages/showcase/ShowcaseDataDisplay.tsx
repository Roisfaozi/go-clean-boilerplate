import { PageHeader } from "@/components/layout/page-header";
import { NexusBadge } from "@casbin/ui";
import {
	DataTableAdvanced,
	type ColumnDef,
} from "@/components/data/data-table-advanced";
import {
	TreeView,
	KeyValueList,
	CodeBlock,
	Timeline,
	ActivityFeed,
	type TreeNode,
	type TimelineItem,
	type ActivityItem,
} from "@/components/data/data-display";
import { Sparkline } from "@/components/charts/charts";
import { Folder, File, FileText } from "lucide-react";

const sampleUsers = Array.from({ length: 25 }, (_, i) => ({
	id: `u${i + 1}`,
	name: `User ${i + 1}`,
	email: `user${i + 1}@nexus.dev`,
	role: ["Admin", "Editor", "Viewer"][i % 3],
	status: ["Active", "Inactive"][i % 2],
}));

const userColumns: ColumnDef<(typeof sampleUsers)[0]>[] = [
	{ id: "name", header: "Name", accessorKey: "name", sortable: true },
	{ id: "email", header: "Email", accessorKey: "email", sortable: true },
	{
		id: "role",
		header: "Role",
		accessorKey: "role",
		sortable: true,
		cell: (row) => <NexusBadge variant="info">{row.role}</NexusBadge>,
	},
	{
		id: "status",
		header: "Status",
		accessorKey: "status",
		cell: (row) => (
			<NexusBadge variant={row.status === "Active" ? "success" : "neutral"}>
				{row.status}
			</NexusBadge>
		),
	},
];

const treeData: TreeNode[] = [
	{
		id: "1",
		label: "src",
		icon: <Folder className="text-primary h-4 w-4" />,
		children: [
			{
				id: "2",
				label: "components",
				icon: <Folder className="text-primary h-4 w-4" />,
				children: [
					{
						id: "3",
						label: "Button.tsx",
						icon: <FileText className="text-muted-foreground h-4 w-4" />,
					},
					{
						id: "4",
						label: "Card.tsx",
						icon: <FileText className="text-muted-foreground h-4 w-4" />,
					},
				],
			},
			{
				id: "5",
				label: "App.tsx",
				icon: <FileText className="text-muted-foreground h-4 w-4" />,
			},
		],
	},
	{
		id: "6",
		label: "package.json",
		icon: <File className="text-muted-foreground h-4 w-4" />,
	},
];

const timelineItems: TimelineItem[] = [
	{
		id: "1",
		title: "Project created",
		description: "NexusOS initialized",
		time: "2 hours ago",
		variant: "success",
	},
	{
		id: "2",
		title: "Team invited",
		description: "3 members added",
		time: "1 hour ago",
		variant: "info",
	},
	{
		id: "3",
		title: "First deployment",
		time: "30 min ago",
		variant: "default",
	},
	{
		id: "4",
		title: "Error detected",
		description: "Build failed on staging",
		time: "10 min ago",
		variant: "danger",
	},
];

const activityItems: ActivityItem[] = [
	{
		id: "1",
		user: { name: "Alice" },
		action: "created project",
		target: "NexusOS",
		time: "5 min ago",
	},
	{
		id: "2",
		user: { name: "Bob" },
		action: "updated role",
		target: "Admin",
		time: "12 min ago",
	},
	{
		id: "3",
		user: { name: "Carol" },
		action: "deleted user",
		target: "test@demo.com",
		time: "1 hour ago",
	},
];

export default function ShowcaseDataDisplay() {
	return (
		<div className="max-w-6xl space-y-10">
			<PageHeader
				title="Data Display"
				description="DataTable, TreeView, KeyValueList, CodeBlock, Timeline, and ActivityFeed."
			/>

			<Section title="Advanced Data Table">
				<DataTableAdvanced
					columns={userColumns}
					data={sampleUsers}
					pageSize={5}
					selectable
					searchable
					bulkActions={[
						{ label: "Delete", onClick: (ids) => console.log("Delete:", ids) },
						{ label: "Export", onClick: (ids) => console.log("Export:", ids) },
					]}
				/>
			</Section>

			<Section title="Tree View & Key-Value List">
				<div className="grid grid-cols-1 gap-6 md:grid-cols-2">
					<div>
						<h3 className="text-h4 text-muted-foreground mb-3">Tree View</h3>
						<div className="border-border rounded-lg border p-3">
							<TreeView
								nodes={treeData}
								onSelect={(n) => console.log("Selected:", n.label)}
							/>
						</div>
					</div>
					<div>
						<h3 className="text-h4 text-muted-foreground mb-3">
							Key-Value List
						</h3>
						<KeyValueList
							striped
							items={[
								{ key: "Name", value: "NexusOS" },
								{ key: "Version", value: "2.1.0" },
								{
									key: "Status",
									value: <NexusBadge variant="success">Active</NexusBadge>,
								},
								{ key: "License", value: "MIT" },
								{
									key: "Build",
									value: <Sparkline data={[3, 7, 4, 8, 5, 9, 6]} />,
								},
							]}
						/>
					</div>
				</div>
			</Section>

			<Section title="Code Block">
				<CodeBlock
					language="typescript"
					code={`import {  NexusButton  } from "@casbin/ui";

export function MyComponent() {
  return <NexusButton variant="primary">Click me</NexusButton>;
}`}
				/>
			</Section>

			<Section title="Timeline & Activity Feed">
				<div className="grid grid-cols-1 gap-6 md:grid-cols-2">
					<div>
						<h3 className="text-h4 text-muted-foreground mb-3">Timeline</h3>
						<Timeline items={timelineItems} />
					</div>
					<div>
						<h3 className="text-h4 text-muted-foreground mb-3">
							Activity Feed
						</h3>
						<ActivityFeed items={activityItems} />
					</div>
				</div>
			</Section>
		</div>
	);
}

function Section({
	title,
	children,
}: {
	title: string;
	children: React.ReactNode;
}) {
	return (
		<section className="space-y-4">
			<h2 className="text-h2 text-foreground">{title}</h2>
			{children}
		</section>
	);
}
