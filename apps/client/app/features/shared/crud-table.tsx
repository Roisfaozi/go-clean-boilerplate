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
  MoreHorizontal,
  Pencil,
  Trash2,
  Eye,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@casbin/ui";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@casbin/ui";

/* ── Types ── */
export interface CrudColumnDef<T> {
  id: string;
  header: string;
  accessorKey?: keyof T;
  cell?: (row: T) => React.ReactNode;
  sortable?: boolean;
  filterable?: boolean;
  filterOptions?: { label: string; value: string }[];
  width?: number;
  minWidth?: number;
}

interface FilterState {
  [columnId: string]: string;
}

interface CrudTableProps<T extends { id: string | number }> {
  columns: CrudColumnDef<T>[];
  data: T[];
  loading?: boolean;
  pageSize?: number;
  onEdit?: (row: T) => void;
  onDelete?: (row: T) => void;
  onView?: (row: T) => void;
  selectable?: boolean;
  bulkActions?: {
    label: string;
    icon?: React.ReactNode;
    onClick: (ids: (string | number)[]) => void;
  }[];
  className?: string;
}

export function CrudTable<T extends { id: string | number }>({
  columns,
  data,
  loading,
  pageSize = 10,
  onEdit,
  onDelete,
  onView,
  selectable,
  bulkActions,
  className,
}: CrudTableProps<T>) {
  const [search, setSearch] = useState("");
  const [sorts, setSorts] = useState<
    { column: string; direction: "asc" | "desc" }[]
  >([]);
  const [filters, setFilters] = useState<FilterState>({});
  const [page, setPage] = useState(0);
  const [selected, setSelected] = useState<Set<string | number>>(new Set());

  const hasActions = onEdit || onDelete || onView;

  // Filter
  const filtered = useMemo(() => {
    let result = data;
    if (search) {
      const lower = search.toLowerCase();
      result = result.filter((row) =>
        columns.some((col) => {
          const val = col.accessorKey ? row[col.accessorKey] : null;
          return val != null && String(val).toLowerCase().includes(lower);
        }),
      );
    }
    Object.entries(filters).forEach(([colId, filterVal]) => {
      if (!filterVal || filterVal === "__all__") return;
      const col = columns.find((c) => c.id === colId);
      if (!col?.accessorKey) return;
      result = result.filter(
        (row) => String(row[col.accessorKey!]) === filterVal,
      );
    });
    return result;
  }, [data, search, filters, columns]);

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
  const totalPages = Math.max(1, Math.ceil(sorted.length / pageSize));
  const paged = sorted.slice(page * pageSize, (page + 1) * pageSize);

  const toggleSort = (colId: string) => {
    setSorts((prev) => {
      const existing = prev.find((s) => s.column === colId);
      if (!existing) return [{ column: colId, direction: "asc" }];
      if (existing.direction === "asc")
        return prev.map((s) =>
          s.column === colId ? { ...s, direction: "desc" as const } : s,
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

  const filterableCols = columns.filter((c) => c.filterable && c.filterOptions);

  return (
    <div className={cn("space-y-4", className)}>
      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-3">
        <div className="relative max-w-sm min-w-[200px] flex-1">
          <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
          <NexusInput
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(0);
            }}
            placeholder="Search…"
            className="h-9 pl-9"
          />
        </div>
        {filterableCols.map((col) => (
          <Select
            key={col.id}
            value={filters[col.id] || "__all__"}
            onValueChange={(val) => {
              setFilters((p) => ({ ...p, [col.id]: val }));
              setPage(0);
            }}
          >
            <SelectTrigger className="h-9 w-[140px]">
              <SelectValue placeholder={col.header} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__all__">All {col.header}</SelectItem>
              {col.filterOptions!.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ))}
        {bulkActions && selected.size > 0 && (
          <div className="ml-auto flex items-center gap-2">
            <span className="text-muted-foreground text-sm">
              {selected.size} selected
            </span>
            {bulkActions.map((action) => (
              <NexusButton
                key={action.label}
                variant="outline"
                size="sm"
                onClick={() => action.onClick(Array.from(selected))}
              >
                {action.icon}
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
            <tr className="border-border bg-muted/30 border-b">
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
                  className="text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase"
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
              {hasActions && <th className="w-14 px-3 py-3" />}
            </tr>
          </thead>
          <tbody>
            {loading ? (
              Array.from({ length: 3 }).map((_, i) => (
                <tr key={i} className="border-border border-b">
                  {selectable && (
                    <td className="px-3 py-4">
                      <div className="bg-muted h-4 w-4 animate-pulse rounded" />
                    </td>
                  )}
                  {columns.map((col) => (
                    <td key={col.id} className="px-4 py-4">
                      <div className="bg-muted h-4 w-24 animate-pulse rounded" />
                    </td>
                  ))}
                  {hasActions && (
                    <td className="px-3 py-4">
                      <div className="bg-muted h-4 w-6 animate-pulse rounded" />
                    </td>
                  )}
                </tr>
              ))
            ) : paged.length === 0 ? (
              <tr>
                <td
                  colSpan={
                    columns.length + (selectable ? 1 : 0) + (hasActions ? 1 : 0)
                  }
                  className="text-muted-foreground py-12 text-center"
                >
                  No records found
                </td>
              </tr>
            ) : (
              paged.map((row) => (
                <tr
                  key={row.id}
                  className={cn(
                    "border-border hover:bg-muted/20 border-b transition-colors last:border-b-0",
                    selected.has(row.id) && "bg-primary/5",
                  )}
                >
                  {selectable && (
                    <td className="px-3 py-3">
                      <Checkbox
                        checked={selected.has(row.id)}
                        onCheckedChange={() => toggleRow(row.id)}
                      />
                    </td>
                  )}
                  {columns.map((col) => (
                    <td key={col.id} className="px-4 py-3 text-sm">
                      {col.cell
                        ? col.cell(row)
                        : col.accessorKey
                          ? String(row[col.accessorKey] ?? "")
                          : ""}
                    </td>
                  ))}
                  {hasActions && (
                    <td className="px-3 py-3">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <NexusButton
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                          >
                            <MoreHorizontal className="h-4 w-4" />
                          </NexusButton>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {onView && (
                            <DropdownMenuItem onClick={() => onView(row)}>
                              <Eye className="mr-2 h-4 w-4" /> View
                            </DropdownMenuItem>
                          )}
                          {onEdit && (
                            <DropdownMenuItem onClick={() => onEdit(row)}>
                              <Pencil className="mr-2 h-4 w-4" /> Edit
                            </DropdownMenuItem>
                          )}
                          {onDelete && (
                            <>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                onClick={() => onDelete(row)}
                                className="text-destructive focus:text-destructive"
                              >
                                <Trash2 className="mr-2 h-4 w-4" /> Delete
                              </DropdownMenuItem>
                            </>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  )}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-between">
        <p className="text-muted-foreground text-xs">
          {sorted.length === 0
            ? "0 records"
            : `${page * pageSize + 1}–${Math.min((page + 1) * pageSize, sorted.length)} of ${sorted.length}`}
        </p>
        <div className="flex items-center gap-1">
          <NexusButton
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => setPage(0)}
            disabled={page === 0}
          >
            <ChevronsLeft className="h-4 w-4" />
          </NexusButton>
          <NexusButton
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => setPage(page - 1)}
            disabled={page === 0}
          >
            <ChevronLeft className="h-4 w-4" />
          </NexusButton>
          <span className="text-foreground px-3 text-sm">
            {page + 1} / {totalPages}
          </span>
          <NexusButton
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => setPage(page + 1)}
            disabled={page >= totalPages - 1}
          >
            <ChevronRight className="h-4 w-4" />
          </NexusButton>
          <NexusButton
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => setPage(totalPages - 1)}
            disabled={page >= totalPages - 1}
          >
            <ChevronsRight className="h-4 w-4" />
          </NexusButton>
        </div>
      </div>
    </div>
  );
}
