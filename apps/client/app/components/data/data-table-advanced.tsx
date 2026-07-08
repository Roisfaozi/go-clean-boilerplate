import { useState, useMemo } from "react";
import { cn } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { Checkbox } from "@casbin/ui";
import {
	ArrowUp,
	ArrowDown,
	ArrowUpDown,
	ChevronLeft,
	ChevronRight,
	ChevronsLeft,
	ChevronsRight,
	Search,
} from "lucide-react";

/* ── Types ── */
export interface ColumnDef<T> {
	id: string;
	header: string;
	accessorKey?: keyof T;
	cell?: (row: T) => React.ReactNode;
	sortable?: boolean;
	pinned?: "left" | "right";
	width?: number;
	minWidth?: number;
}

interface SortState {
	column: string;
	direction: "asc" | "desc";
}

interface DataTableAdvancedProps<T extends { id: string | number }> {
	columns: ColumnDef<T>[];
	data: T[];
	pageSize?: number;
	selectable?: boolean;
	bulkActions?: {
		label: string;
		onClick: (selectedIds: (string | number)[]) => void;
	}[];
	searchable?: boolean;
	className?: string;
}

export function DataTableAdvanced<T extends { id: string | number }>({
	columns,
	data,
	pageSize = 10,
	selectable,
	bulkActions,
	searchable,
	className,
}: DataTableAdvancedProps<T>) {
	const [search, setSearch] = useState("");
	const [sorts, setSorts] = useState<SortState[]>([]);
	const [page, setPage] = useState(0);
	const [selected, setSelected] = useState<Set<string | number>>(new Set());

	// Filter
	const filtered = useMemo(() => {
		if (!search) return data;
		const lower = search.toLowerCase();
		return data.filter((row) =>
			columns.some((col) => {
				const val = col.accessorKey ? row[col.accessorKey] : null;
				return val != null && String(val).toLowerCase().includes(lower);
			}),
		);
	}, [data, search, columns]);

	// Sort
	const sorted = useMemo(() => {
		if (sorts.length === 0) return filtered;
		return [...filtered].sort((a, b) => {
			for (const sort of sorts) {
				const col = columns.find((c) => c.id === sort.column);
				if (!col?.accessorKey) continue;
				const aVal = a[col.accessorKey];
				const bVal = b[col.accessorKey];
				const cmp = String(aVal ?? "").localeCompare(
					String(bVal ?? ""),
					undefined,
					{
						numeric: true,
					},
				);
				if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
			}
			return 0;
		});
	}, [filtered, sorts, columns]);

	// Paginate
	const totalPages = Math.ceil(sorted.length / pageSize);
	const paged = sorted.slice(page * pageSize, (page + 1) * pageSize);

	const toggleSort = (colId: string) => {
		setSorts((prev) => {
			const existing = prev.find((s) => s.column === colId);
			if (!existing) return [...prev, { column: colId, direction: "asc" }];
			if (existing.direction === "asc")
				return prev.map((s) =>
					s.column === colId ? { ...s, direction: "desc" } : s,
				);
			return prev.filter((s) => s.column !== colId);
		});
	};

	const getSortIcon = (colId: string) => {
		const s = sorts.find((s) => s.column === colId);
		if (!s)
			return <ArrowUpDown className="text-muted-foreground h-3.5 w-3.5" />;
		return s.direction === "asc" ? (
			<ArrowUp className="text-primary h-3.5 w-3.5" />
		) : (
			<ArrowDown className="text-primary h-3.5 w-3.5" />
		);
	};

	const toggleAll = () => {
		if (selected.size === paged.length) setSelected(new Set());
		else setSelected(new Set(paged.map((r) => r.id)));
	};

	const toggleRow = (id: string | number) => {
		const next = new Set(selected);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		setSelected(next);
	};

	return (
		<div className={cn("space-y-4", className)}>
			{/* Toolbar */}
			<div className="flex items-center justify-between gap-3">
				{searchable && (
					<div className="relative max-w-sm">
						<Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
						<NexusInput
							value={search}
							onChange={(e) => {
								setSearch(e.target.value);
								setPage(0);
							}}
							placeholder="Search…"
							className="pl-9"
						/>
					</div>
				)}
				{bulkActions && selected.size > 0 && (
					<div className="flex items-center gap-2">
						<span className="text-small text-muted-foreground">
							{selected.size} selected
						</span>
						{bulkActions.map((action) => (
							<NexusButton
								key={action.label}
								variant="outline"
								size="sm"
								onClick={() => action.onClick(Array.from(selected))}
							>
								{action.label}
							</NexusButton>
						))}
					</div>
				)}
			</div>

			{/* Table */}
			<div className="border-border overflow-auto rounded-lg border">
				<table className="w-full">
					<thead>
						<tr className="border-border bg-surface border-b">
							{selectable && (
								<th className="w-10 px-3 py-3">
									<Checkbox
										checked={paged.length > 0 && selected.size === paged.length}
										onCheckedChange={toggleAll}
									/>
								</th>
							)}
							{columns.map((col) => (
								<th
									key={col.id}
									className="text-caption text-muted-foreground px-[var(--table-cell-padding)] py-3 text-left font-semibold"
									style={{ width: col.width, minWidth: col.minWidth }}
								>
									{col.sortable ? (
										<button
											onClick={() => toggleSort(col.id)}
											className="hover:text-foreground inline-flex items-center gap-1.5 transition-colors"
										>
											{col.header}
											{getSortIcon(col.id)}
										</button>
									) : (
										col.header
									)}
								</th>
							))}
						</tr>
					</thead>
					<tbody>
						{paged.length === 0 ? (
							<tr>
								<td
									colSpan={columns.length + (selectable ? 1 : 0)}
									className="text-muted-foreground text-body py-12 text-center"
								>
									No data found
								</td>
							</tr>
						) : (
							paged.map((row) => (
								<tr
									key={row.id}
									className={cn(
										"border-border hover:bg-surface-hover h-table-row border-b transition-colors last:border-b-0",
										selected.has(row.id) && "bg-primary/5",
									)}
								>
									{selectable && (
										<td className="px-3">
											<Checkbox
												checked={selected.has(row.id)}
												onCheckedChange={() => toggleRow(row.id)}
											/>
										</td>
									)}
									{columns.map((col) => (
										<td
											key={col.id}
											className="text-body px-[var(--table-cell-padding)] py-3"
										>
											{col.cell
												? col.cell(row)
												: col.accessorKey
													? String(row[col.accessorKey] ?? "")
													: ""}
										</td>
									))}
								</tr>
							))
						)}
					</tbody>
				</table>
			</div>

			{/* Pagination */}
			{totalPages > 1 && (
				<div className="flex items-center justify-between">
					<p className="text-caption text-muted-foreground">
						{page * pageSize + 1}–
						{Math.min((page + 1) * pageSize, sorted.length)} of {sorted.length}
					</p>
					<div className="flex items-center gap-1">
						<NexusButton
							variant="outline"
							size="icon"
							onClick={() => setPage(0)}
							disabled={page === 0}
						>
							<ChevronsLeft className="h-4 w-4" />
						</NexusButton>
						<NexusButton
							variant="outline"
							size="icon"
							onClick={() => setPage(page - 1)}
							disabled={page === 0}
						>
							<ChevronLeft className="h-4 w-4" />
						</NexusButton>
						<span className="text-small text-foreground px-3">
							{page + 1} / {totalPages}
						</span>
						<NexusButton
							variant="outline"
							size="icon"
							onClick={() => setPage(page + 1)}
							disabled={page >= totalPages - 1}
						>
							<ChevronRight className="h-4 w-4" />
						</NexusButton>
						<NexusButton
							variant="outline"
							size="icon"
							onClick={() => setPage(totalPages - 1)}
							disabled={page >= totalPages - 1}
						>
							<ChevronsRight className="h-4 w-4" />
						</NexusButton>
					</div>
				</div>
			)}
		</div>
	);
}
