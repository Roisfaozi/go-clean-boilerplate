import { useEffect, useMemo, useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { useAuditLogs, useExportAuditLogs } from "./auditHooks";
import {
	AuditLogTable,
	type AuditLogEntry,
} from "@/components/admin/audit-log-table";
import { NexusButton, NexusCard, NexusInput, Skeleton } from "@casbin/ui";
import { format } from "date-fns";
import { Download, RefreshCw, Search } from "lucide-react";

export default function AuditLogsPage() {
	const [page, setPage] = useState(1);
	const [search, setSearch] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const limit = 15;

	useEffect(() => {
		const timeout = setTimeout(() => {
			setDebouncedSearch(search);
		}, 300);
		return () => clearTimeout(timeout);
	}, [search]);

	const {
		data: response,
		isLoading,
		refetch,
		isFetching,
	} = useAuditLogs({
		page,
		limit,
		search: debouncedSearch || undefined,
		sort: "created_at desc",
	});

	const exportLogs = useExportAuditLogs();

	const logs: AuditLogEntry[] = useMemo(() => {
		return (
			response?.data?.map((log) => ({
				id: log.id,
				action: log.action,
				actor: log.user_id,
				target: `${log.entity}:${log.entity_id}`,
				ip_address: log.ip_address,
				timestamp: format(
					new Date(log.created_at * 1000),
					"yyyy-MM-dd HH:mm:ss",
				),
				severity:
					log.action.includes("delete") ||
					log.action.includes("failed") ||
					log.action.includes("revoke")
						? "critical"
						: log.action.includes("update") ||
								log.action.includes("permission") ||
								log.action.includes("grant")
							? "warning"
							: "info",
			})) || []
		);
	}, [response]);

	const totalPages = Math.ceil((response?.meta?.total || 0) / limit);

	return (
		<div className="space-y-6">
			<PageHeader
				title="Audit Logs"
				description="System-wide activity trail and change history."
				actions={
					<div className="flex gap-2">
						<NexusButton
							variant="outline"
							size="sm"
							onClick={() => exportLogs({ format: "csv" })}
						>
							<Download className="mr-2 h-4 w-4" /> Export CSV
						</NexusButton>
						<NexusButton
							variant="outline"
							size="sm"
							onClick={() => refetch()}
							disabled={isFetching}
						>
							<RefreshCw
								className={`mr-2 h-4 w-4 ${isFetching ? "animate-spin" : ""}`}
							/>
							Refresh
						</NexusButton>
					</div>
				}
			/>

			<NexusCard className="shadow-premium overflow-hidden border-none bg-white/50 p-0 backdrop-blur-sm">
				<div className="bg-muted/30 border-b p-4">
					<div className="relative max-w-md">
						<Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
						<NexusInput
							placeholder="Search by action, user, or target..."
							value={search}
							onChange={(e) => {
								setSearch(e.target.value);
								setPage(1);
							}}
							className="bg-white pl-9"
						/>
					</div>
				</div>

				{isLoading ? (
					<div className="space-y-4 p-6">
						{Array.from({ length: 8 }).map((_, i) => (
							<Skeleton key={i} className="h-12 w-full rounded-lg" />
						))}
					</div>
				) : (
					<AuditLogTable
						logs={logs}
						// We'll update the table component to handle server-side pagination if needed,
						// but for now it has its own internal pagination which we'll suppress by passing
						// only the current page's data.
					/>
				)}

				<div className="bg-muted/10 flex items-center justify-between border-t p-4">
					<span className="text-muted-foreground text-sm">
						Total Records: {response?.meta?.total || 0}
					</span>
					<div className="flex gap-2">
						<NexusButton
							variant="outline"
							size="sm"
							disabled={page <= 1}
							onClick={() => setPage((p) => p - 1)}
						>
							Previous
						</NexusButton>
						<div className="flex items-center px-4 text-sm font-medium">
							Page {page} of {totalPages || 1}
						</div>
						<NexusButton
							variant="outline"
							size="sm"
							disabled={page >= totalPages}
							onClick={() => setPage((p) => p + 1)}
						>
							Next
						</NexusButton>
					</div>
				</div>
			</NexusCard>
		</div>
	);
}
