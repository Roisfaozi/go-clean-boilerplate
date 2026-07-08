import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { NexusButton } from "@casbin/ui";
import {
	NexusCard,
	NexusCardHeader,
	NexusCardTitle,
	NexusCardDescription,
	NexusCardContent,
} from "@casbin/ui";
import { RealtimeIndicator } from "@/components/realtime/realtime-indicator";
import { LiveActivityFeed } from "@/components/realtime/live-activity-feed";
import { PresenceAvatars } from "@/components/realtime/presence-avatars";
import { NotificationBell } from "@/components/realtime/notification-bell";
import {
	useNotificationStore,
	useActivityStore,
	usePresenceStore,
} from "@/stores/realtime-store";
import { Users, Activity, Bell } from "lucide-react";

export default function ShowcaseRealtime() {
	const addNotification = useNotificationStore((s) => s.addNotification);
	const addActivity = useActivityStore((s) => s.addActivity);
	const addUser = usePresenceStore((s) => s.addUser);
	const removeUser = usePresenceStore((s) => s.removeUser);
	const [counter, setCounter] = useState(0);

	const simulateNotification = () => {
		const types = ["info", "success", "warning", "danger"] as const;
		const titles = [
			"New user registered",
			"Deployment successful",
			"High memory usage",
			"Authentication failure detected",
		];
		const idx = counter % 4;
		addNotification({
			title: titles[idx],
			description: "Simulated realtime event",
			type: types[idx],
		});
		setCounter((c) => c + 1);
	};

	const simulateActivity = () => {
		const actions = [
			{
				user: "Alice",
				action: "pushed to",
				target: "main branch",
				type: "success" as const,
			},
			{
				user: "Bob",
				action: "updated",
				target: "user permissions",
				type: "info" as const,
			},
			{
				user: "System",
				action: "detected",
				target: "memory spike",
				type: "warning" as const,
			},
			{
				user: "Carol",
				action: "deleted",
				target: "staging environment",
				type: "danger" as const,
			},
		];
		addActivity(actions[counter % 4]);
		setCounter((c) => c + 1);
	};

	const simulatePresence = () => {
		const id = `sim-${Date.now()}`;
		const names = ["Frank", "Grace", "Hank", "Ivy", "Jack"];
		addUser({ id, name: names[counter % 5], status: "online" });
		setCounter((c) => c + 1);
		// Auto remove after 5s
		setTimeout(() => removeUser(id), 5000);
	};

	return (
		<div className="max-w-6xl space-y-10">
			<PageHeader
				title="Realtime"
				description="SSE & WebSocket powered live components — notifications, presence, and activity feed."
			/>

			{/* Simulate controls */}
			<NexusCard>
				<NexusCardHeader>
					<NexusCardTitle>Simulate Events</NexusCardTitle>
					<NexusCardDescription>
						Trigger mock events to see the components update in realtime.
					</NexusCardDescription>
				</NexusCardHeader>
				<NexusCardContent>
					<div className="flex flex-wrap gap-3">
						<NexusButton variant="primary" onClick={simulateNotification}>
							<Bell className="mr-1 h-4 w-4" /> Add Notification
						</NexusButton>
						<NexusButton variant="secondary" onClick={simulateActivity}>
							<Activity className="mr-1 h-4 w-4" /> Add Activity
						</NexusButton>
						<NexusButton variant="outline" onClick={simulatePresence}>
							<Users className="mr-1 h-4 w-4" /> Simulate User Join (5s)
						</NexusButton>
					</div>
				</NexusCardContent>
			</NexusCard>

			<div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
				{/* Realtime Indicator */}
				<NexusCard>
					<NexusCardHeader>
						<NexusCardTitle>RealtimeIndicator</NexusCardTitle>
						<NexusCardDescription>
							Shows SSE & WebSocket connection status.
						</NexusCardDescription>
					</NexusCardHeader>
					<NexusCardContent className="space-y-4">
						<div className="flex items-center gap-6">
							<div className="flex items-center gap-2">
								<span className="text-small text-muted-foreground">
									Compact:
								</span>
								<RealtimeIndicator />
							</div>
							<div className="flex items-center gap-2">
								<span className="text-small text-muted-foreground">
									With label:
								</span>
								<RealtimeIndicator showLabel />
							</div>
						</div>
						<p className="text-caption text-muted-foreground">
							Hover to see detailed SSE/WS status. In demo mode, connections
							show as offline since there's no backend.
						</p>
					</NexusCardContent>
				</NexusCard>

				{/* Notification Bell */}
				<NexusCard>
					<NexusCardHeader>
						<NexusCardTitle>NotificationBell</NexusCardTitle>
						<NexusCardDescription>
							Dropdown notification center with unread badge.
						</NexusCardDescription>
					</NexusCardHeader>
					<NexusCardContent>
						<div className="flex items-center gap-4">
							<NotificationBell />
							<span className="text-small text-muted-foreground">
								← Click the bell icon
							</span>
						</div>
					</NexusCardContent>
				</NexusCard>

				{/* Presence Avatars */}
				<NexusCard>
					<NexusCardHeader>
						<NexusCardTitle>PresenceAvatars</NexusCardTitle>
						<NexusCardDescription>
							Shows online users with status indicators.
						</NexusCardDescription>
					</NexusCardHeader>
					<NexusCardContent className="space-y-4">
						<div className="space-y-3">
							<div className="flex items-center gap-3">
								<span className="text-small text-muted-foreground w-12">
									Small:
								</span>
								<PresenceAvatars size="sm" max={4} />
							</div>
							<div className="flex items-center gap-3">
								<span className="text-small text-muted-foreground w-12">
									Medium:
								</span>
								<PresenceAvatars size="md" max={4} />
							</div>
							<div className="flex items-center gap-3">
								<span className="text-small text-muted-foreground w-12">
									Large:
								</span>
								<PresenceAvatars size="lg" max={3} />
							</div>
						</div>
					</NexusCardContent>
				</NexusCard>

				{/* Live Activity Feed */}
				<NexusCard>
					<NexusCardHeader>
						<NexusCardTitle>LiveActivityFeed</NexusCardTitle>
						<NexusCardDescription>
							Realtime activity stream with typed events.
						</NexusCardDescription>
					</NexusCardHeader>
					<NexusCardContent>
						<LiveActivityFeed maxItems={6} />
					</NexusCardContent>
				</NexusCard>
			</div>
		</div>
	);
}
