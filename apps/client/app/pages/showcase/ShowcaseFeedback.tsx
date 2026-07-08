import { PageHeader } from "@/components/layout/page-header";
import { Spinner } from "@casbin/ui";
import {
	InlineAlert,
	StatusIndicator,
	ProgressBar,
	SkeletonLoader,
	NotificationCenter,
	type Notification,
} from "@casbin/ui";
import { User, Code, ShieldCheck } from "lucide-react";

const notifications: Notification[] = [
	{
		id: "1",
		title: "New user registered",
		description: "alice@nexus.dev signed up",
		time: "2m",
		icon: <User className="text-primary h-4 w-4" />,
	},
	{
		id: "2",
		title: "Deployment complete",
		description: "v2.1.0 live",
		time: "15m",
		read: true,
		icon: <Code className="text-success h-4 w-4" />,
	},
	{
		id: "3",
		title: "Security alert",
		description: "Unusual login detected",
		time: "1h",
		icon: <ShieldCheck className="text-danger h-4 w-4" />,
	},
];

export default function ShowcaseFeedback() {
	return (
		<div className="max-w-5xl space-y-10">
			<PageHeader
				title="Feedback"
				description="Alerts, status indicators, progress bars, skeletons, notifications, and spinners."
			/>

			<Section title="Inline Alerts">
				<div className="space-y-3">
					<InlineAlert variant="info" title="Info">
						This is an informational message.
					</InlineAlert>
					<InlineAlert variant="success" title="Success">
						Operation completed successfully.
					</InlineAlert>
					<InlineAlert variant="warning" title="Warning" dismissible>
						This can be dismissed.
					</InlineAlert>
					<InlineAlert variant="danger" title="Error">
						Something went wrong.
					</InlineAlert>
				</div>
			</Section>

			<Section title="Status Indicators">
				<div className="flex flex-wrap gap-6">
					<StatusIndicator status="online" label="Online" />
					<StatusIndicator status="away" label="Away" />
					<StatusIndicator status="busy" label="Busy" />
					<StatusIndicator status="offline" label="Offline" />
				</div>
			</Section>

			<Section title="Progress Bars">
				<div className="max-w-lg space-y-4">
					<ProgressBar value={75} label="Upload" showValue variant="primary" />
					<ProgressBar
						value={45}
						label="Processing"
						showValue
						variant="info"
						size="sm"
					/>
					<ProgressBar
						value={90}
						label="Complete"
						showValue
						variant="success"
						size="lg"
					/>
					<ProgressBar value={30} variant="warning" showValue />
				</div>
			</Section>

			<Section title="Skeleton Loaders">
				<div className="grid grid-cols-1 gap-4 md:grid-cols-3">
					<div>
						<p className="text-caption text-muted-foreground mb-2">Text</p>
						<SkeletonLoader variant="text" />
					</div>
					<div>
						<p className="text-caption text-muted-foreground mb-2">Card</p>
						<SkeletonLoader variant="card" />
					</div>
					<div>
						<p className="text-caption text-muted-foreground mb-2">Table</p>
						<SkeletonLoader variant="table" rows={3} />
					</div>
				</div>
			</Section>

			<Section title="Notification Center">
				<NotificationCenter notifications={notifications} />
			</Section>

			<Section title="Spinner">
				<div className="flex items-center gap-6">
					<Spinner size="sm" />
					<Spinner size="md" />
					<Spinner size="lg" />
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
