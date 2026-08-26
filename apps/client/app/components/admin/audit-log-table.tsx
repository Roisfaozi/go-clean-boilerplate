import { Badge } from "@casbin/ui";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@casbin/ui";
import { Search } from "lucide-react";

export interface AuditLogEntry {
	id: string;
	action: string;
	actor: string;
	target: string;
	ip_address: string;
	timestamp: string;
	severity: "info" | "warning" | "critical";
	details?: string;
}

const severityMap: Record<
	string,
	"default" | "secondary" | "destructive" | "outline"
> = {
	info: "secondary",
	warning: "outline",
	critical: "destructive",
};

const severityDot: Record<string, string> = {
	info: "bg-info",
	warning: "bg-warning",
	critical: "bg-destructive",
};

interface AuditLogTableProps {
	logs: AuditLogEntry[];
}

export function AuditLogTable({ logs }: AuditLogTableProps) {
	return (
		<div className="border-border overflow-hidden border-t">
			<Table>
				<TableHeader>
					<TableRow className="bg-muted/50 border-none">
						<TableHead className="pl-8">Action</TableHead>
						<TableHead>Actor</TableHead>
						<TableHead>Target</TableHead>
						<TableHead>Severity</TableHead>
						<TableHead>IP Address</TableHead>
						<TableHead className="pr-8">Timestamp</TableHead>
					</TableRow>
				</TableHeader>
				<TableBody>
					{logs.length === 0 ? (
						<TableRow>
							<TableCell
								colSpan={6}
								className="text-muted-foreground py-12 text-center"
							>
								<div className="flex flex-col items-center gap-2">
									<Search className="text-muted-foreground/20 h-8 w-8" />
									<p>No activity logs found matching your criteria</p>
								</div>
							</TableCell>
						</TableRow>
					) : (
						logs.map((log) => (
							<TableRow
								key={log.id}
								className="hover:bg-muted/30 group transition-colors"
							>
								<TableCell className="text-foreground/70 pl-8 font-mono text-[10px] font-bold tracking-wider uppercase">
									{log.action}
								</TableCell>
								<TableCell className="text-foreground font-medium">
									{log.actor}
								</TableCell>
								<TableCell className="text-muted-foreground text-xs">
									{log.target}
								</TableCell>
								<TableCell>
									<Badge
										variant={severityMap[log.severity] ?? "default"}
										className="h-6 gap-1.5"
									>
										<span
											className={`h-1.5 w-1.5 rounded-full ${severityDot[log.severity]}`}
										/>
										{log.severity}
									</Badge>
								</TableCell>
								<TableCell className="text-muted-foreground font-mono text-[10px]">
									{log.ip_address}
								</TableCell>
								<TableCell className="text-muted-foreground pr-8 text-xs whitespace-nowrap">
									{log.timestamp}
								</TableCell>
							</TableRow>
						))
					)}
				</TableBody>
			</Table>
		</div>
	);
}
