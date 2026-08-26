"use client";

import { useDashboard } from "./dashboard-context";
import { Badge } from "@casbin/ui";
import { ActivityChart } from "~/components/dashboard/activity-chart";

export function SystemInsights() {
	const { insights } = useDashboard();

	return (
		<div className="grid grid-cols-1 gap-gap lg:grid-cols-2">
			<ActivityChart />
			<div className="bg-card text-card-foreground rounded-lg border p-6 shadow-sm">
				<div className="mb-4 flex items-center justify-between">
					<h2 className="text-primary text-lg font-semibold tracking-tight">
						System Insights
					</h2>
					<Badge variant="outline" className="bg-primary/5">
						Live
					</Badge>
				</div>
				<div className="space-y-4">
					<div className="bg-muted/30 rounded-lg border border-dashed p-4">
						<p className="text-muted-foreground text-sm leading-relaxed italic">
							{insights
								? `User engagement is active. Most active role in audit log: ${insights.most_active_role}.`
								: "Analyzing audit activity patterns..."}
						</p>
					</div>
				</div>
			</div>
		</div>
	);
}
