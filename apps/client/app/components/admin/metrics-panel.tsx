import {
	NexusCard,
	NexusCardHeader,
	NexusCardTitle,
	NexusCardDescription,
} from "@casbin/ui";
import { cn } from "@casbin/ui";
import {
	AreaChart,
	Area,
	BarChart,
	Bar,
	LineChart,
	Line,
	XAxis,
	YAxis,
	CartesianGrid,
	Tooltip,
	ResponsiveContainer,
} from "recharts";
import { TrendingUp, TrendingDown, Minus, type LucideIcon } from "lucide-react";

export interface MetricSummary {
	label: string;
	value: string;
	change: number;
	changeLabel: string;
	icon: LucideIcon;
}

export interface TimeSeriesPoint {
	label: string;
	value: number;
	value2?: number;
}

const tooltipStyle = {
	background: "hsl(var(--popover))",
	border: "1px solid hsl(var(--border))",
	borderRadius: 8,
	fontSize: 13,
};

function TrendBadge({ value }: { value: number }) {
	const Icon = value > 0 ? TrendingUp : value < 0 ? TrendingDown : Minus;
	const color =
		value > 0
			? "text-success"
			: value < 0
				? "text-destructive"
				: "text-muted-foreground";
	return (
		<span
			className={cn(
				"inline-flex items-center gap-1 text-xs font-medium",
				color,
			)}
		>
			<Icon className="h-3 w-3" />
			{Math.abs(value)}%
		</span>
	);
}

interface MetricsPanelProps {
	title: string;
	description?: string;
	summaries: MetricSummary[];
	chartData: TimeSeriesPoint[];
	chartType?: "area" | "bar" | "line";
	chartLabel?: string;
	chartLabel2?: string;
}

export function MetricsPanel({
	title,
	description,
	summaries,
	chartData,
	chartType = "area",
	chartLabel = "Value",
	chartLabel2,
}: MetricsPanelProps) {
	return (
		<NexusCard>
			<NexusCardHeader>
				<NexusCardTitle>{title}</NexusCardTitle>
				{description && (
					<NexusCardDescription>{description}</NexusCardDescription>
				)}
			</NexusCardHeader>

			{/* Summary cards */}
			<div className="mb-6 grid grid-cols-2 gap-3 md:grid-cols-4">
				{summaries.map((m) => (
					<div key={m.label} className="bg-muted/50 space-y-1 rounded-lg p-3">
						<div className="flex items-center justify-between">
							<m.icon className="text-muted-foreground h-4 w-4" />
							<TrendBadge value={m.change} />
						</div>
						<p className="text-foreground text-xl font-bold">{m.value}</p>
						<p className="text-muted-foreground text-xs">{m.label}</p>
					</div>
				))}
			</div>

			{/* Chart */}
			<div className="h-[280px]">
				<ResponsiveContainer width="100%" height="100%">
					{chartType === "bar" ? (
						<BarChart data={chartData}>
							<CartesianGrid
								strokeDasharray="3 3"
								stroke="hsl(var(--border))"
							/>
							<XAxis
								dataKey="label"
								tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
							/>
							<YAxis
								tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
							/>
							<Tooltip contentStyle={tooltipStyle} />
							<Bar
								dataKey="value"
								name={chartLabel}
								fill="hsl(var(--primary))"
								radius={[4, 4, 0, 0]}
							/>
							{chartLabel2 && (
								<Bar
									dataKey="value2"
									name={chartLabel2}
									fill="hsl(var(--secondary))"
									radius={[4, 4, 0, 0]}
								/>
							)}
						</BarChart>
					) : chartType === "line" ? (
						<LineChart data={chartData}>
							<CartesianGrid
								strokeDasharray="3 3"
								stroke="hsl(var(--border))"
							/>
							<XAxis
								dataKey="label"
								tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
							/>
							<YAxis
								tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
							/>
							<Tooltip contentStyle={tooltipStyle} />
							<Line
								type="monotone"
								dataKey="value"
								name={chartLabel}
								stroke="hsl(var(--primary))"
								strokeWidth={2}
								dot={false}
							/>
							{chartLabel2 && (
								<Line
									type="monotone"
									dataKey="value2"
									name={chartLabel2}
									stroke="hsl(var(--destructive))"
									strokeWidth={2}
									dot={false}
									strokeDasharray="5 5"
								/>
							)}
						</LineChart>
					) : (
						<AreaChart data={chartData}>
							<defs>
								<linearGradient id="metricsGrad" x1="0" y1="0" x2="0" y2="1">
									<stop
										offset="0%"
										stopColor="hsl(var(--primary))"
										stopOpacity={0.2}
									/>
									<stop
										offset="100%"
										stopColor="hsl(var(--primary))"
										stopOpacity={0}
									/>
								</linearGradient>
							</defs>
							<CartesianGrid
								strokeDasharray="3 3"
								stroke="hsl(var(--border))"
							/>
							<XAxis
								dataKey="label"
								tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
							/>
							<YAxis
								tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
							/>
							<Tooltip contentStyle={tooltipStyle} />
							<Area
								type="monotone"
								dataKey="value"
								name={chartLabel}
								stroke="hsl(var(--primary))"
								fill="url(#metricsGrad)"
								strokeWidth={2}
							/>
						</AreaChart>
					)}
				</ResponsiveContainer>
			</div>
		</NexusCard>
	);
}
