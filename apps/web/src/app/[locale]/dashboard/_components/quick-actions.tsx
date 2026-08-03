"use client";

import { Button } from "@casbin/ui";
import Link from "next/link";
import { useDensity } from "~/components/shared/providers/density-provider";
import { Icon } from "~/components/shared/icon";
import { auditApi } from "~/lib/api/audit";

export function QuickActions() {
	const { density } = useDensity();
	const isCompact = density === "compact";
	const actionClassName = isCompact
		? "h-auto min-h-12 w-full justify-start rounded-lg px-3 py-3 text-left [&_svg]:size-4"
		: "h-auto min-h-16 w-full justify-start rounded-lg px-4 py-4 text-left [&_svg]:size-5";
	const iconBoxClassName = isCompact ? "rounded-md p-2" : "rounded-md p-2.5";
	const titleClassName = isCompact
		? "text-xs font-semibold"
		: "text-sm font-semibold";
	const descriptionClassName = isCompact
		? "text-muted-foreground text-xs"
		: "text-muted-foreground text-sm";

	return (
		<div className="flex flex-col gap-4 md:col-span-2">
			<h2 className="text-lg font-semibold tracking-tight">Quick Actions</h2>
			<div className="grid gap-3">
				<Link href="/dashboard/users" className="block w-full">
					<Button className={actionClassName} variant="outline" size="lg">
						<div className="flex items-center gap-3">
							<div className={`bg-primary/10 text-primary ${iconBoxClassName}`}>
								<Icon name="UserPlus" className="h-5 w-5" />
							</div>
							<div className="text-left">
								<div className={titleClassName}>Manage Users</div>
								<div className={descriptionClassName}>Add or edit accounts</div>
							</div>
						</div>
					</Button>
				</Link>

				<Link href="/dashboard/roles" className="block w-full">
					<Button className={actionClassName} variant="outline" size="lg">
						<div className="flex items-center gap-3">
							<div className={`bg-accent/10 text-accent ${iconBoxClassName}`}>
								<Icon name="Shield" className="h-5 w-5" />
							</div>
							<div className="text-left">
								<div className={titleClassName}>Configure Roles</div>
								<div className={descriptionClassName}>Update permissions</div>
							</div>
						</div>
					</Button>
				</Link>

				<Button
					className={actionClassName}
					variant="outline"
					size="lg"
					onClick={() => window.open(auditApi.export(), "_blank")}
				>
					<div className="flex items-center gap-3">
						<div
							className={`bg-secondary/10 text-secondary ${iconBoxClassName}`}
						>
							<Icon name="Download" className="h-5 w-5" />
						</div>
						<div className="text-left">
							<div className={titleClassName}>Export Logs</div>
							<div className={descriptionClassName}>Download audit trail</div>
						</div>
					</div>
				</Button>
			</div>
		</div>
	);
}
