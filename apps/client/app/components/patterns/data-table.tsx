import type * as React from "react";
import { cn } from "@casbin/ui";
import { Spinner } from "@casbin/ui";
import { ChevronUp, ChevronDown, ChevronsUpDown } from "lucide-react";

// ─── DataTable ───
interface Column<T> {
	key: string;
	header: string;
	sortable?: boolean;
	render?: (item: T) => React.ReactNode;
	className?: string;
}

interface DataTableProps<T> {
	columns: Column<T>[];
	data: T[];
	loading?: boolean;
	emptyMessage?: string;
	onSort?: (key: string, direction: "asc" | "desc") => void;
	sortKey?: string;
	sortDirection?: "asc" | "desc";
	className?: string;
}

export function DataTable<T extends Record<string, unknown>>({
	columns,
	data,
	loading,
	emptyMessage = "No data found",
	onSort,
	sortKey,
	sortDirection,
	className,
}: DataTableProps<T>) {
	const handleSort = (key: string) => {
		if (!onSort) return;
		const newDir = sortKey === key && sortDirection === "asc" ? "desc" : "asc";
		onSort(key, newDir);
	};

	return (
		<div
			className={cn("border-border overflow-auto rounded-lg border", className)}
		>
			<table className="text-body w-full">
				<thead>
					<tr className="border-border bg-surface border-b">
						{columns.map((col) => (
							<th
								key={col.key}
								className={cn(
									"h-table-row text-muted-foreground px-[var(--table-cell-padding)] text-left font-medium",
									col.sortable &&
										"hover:text-foreground cursor-pointer select-none",
									col.className,
								)}
								onClick={() => col.sortable && handleSort(col.key)}
							>
								<div className="flex items-center gap-1">
									{col.header}
									{col.sortable &&
										(sortKey === col.key ? (
											sortDirection === "asc" ? (
												<ChevronUp className="h-4 w-4" />
											) : (
												<ChevronDown className="h-4 w-4" />
											)
										) : (
											<ChevronsUpDown className="h-3.5 w-3.5 opacity-40" />
										))}
								</div>
							</th>
						))}
					</tr>
				</thead>
				<tbody>
					{loading ? (
						<tr>
							<td colSpan={columns.length} className="h-32 text-center">
								<Spinner size="lg" />
							</td>
						</tr>
					) : data.length === 0 ? (
						<tr>
							<td
								colSpan={columns.length}
								className="text-muted-foreground h-32 text-center"
							>
								{emptyMessage}
							</td>
						</tr>
					) : (
						data.map((item, idx) => (
							<tr
								key={idx}
								className="border-border hover:bg-surface-hover border-b transition-colors last:border-0"
							>
								{columns.map((col) => (
									<td
										key={col.key}
										className={cn(
											"h-table-row px-[var(--table-cell-padding)]",
											col.className,
										)}
									>
										{col.render
											? col.render(item)
											: String(item[col.key] ?? "")}
									</td>
								))}
							</tr>
						))
					)}
				</tbody>
			</table>
		</div>
	);
}
