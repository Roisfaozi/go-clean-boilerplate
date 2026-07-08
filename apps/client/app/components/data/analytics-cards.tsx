import { cn } from "@casbin/ui";
import { TrendingUp, TrendingDown, Minus, type LucideIcon } from "lucide-react";
import { ResponsiveContainer } from "recharts";

/* ═══════ TrendIndicator ═══════ */
interface TrendIndicatorProps {
	value: number;
	label?: string;
	size?: "sm" | "md";
	className?: string;
}

export function TrendIndicator({
	value,
	label,
	size = "md",
	className,
}: TrendIndicatorProps) {
	const isPositive = value > 0;
	const isNeutral = value === 0;
	const Icon = isNeutral ? Minus : isPositive ? TrendingUp : TrendingDown;
	const color = isNeutral
		? "text-muted-foreground"
		: isPositive
			? "text-success"
			: "text-danger";
	const textSize = size === "sm" ? "text-caption" : "text-small";

	return (
		<span className={cn("inline-flex items-center gap-1", color, className)}>
			<Icon className={cn(size === "sm" ? "h-3 w-3" : "h-4 w-4")} />
			<span className={cn(textSize, "font-medium")}>
				{isPositive ? "+" : ""}
				{value}%
			</span>
			{label && (
				<span className={cn(textSize, "text-muted-foreground")}>{label}</span>
			)}
		</span>
	);
}

/* ═══════ MetricCard ═══════ */
interface MetricCardProps {
	title: string;
	value: string | number;
	trend?: { value: number; label?: string };
	icon?: LucideIcon;
	iconColor?: string;
	sparkline?: React.ReactNode;
	className?: string;
}

export function MetricCard({
	title,
	value,
	trend,
	icon: Icon,
	iconColor,
	sparkline,
	className,
}: MetricCardProps) {
	return (
		<div
			className={cn(
				"bg-card border-border p-card-pad rounded-lg border shadow-sm",
				className,
			)}
		>
			<div className="flex items-start justify-between">
				<div className="flex-1 space-y-1">
					<p className="text-small text-muted-foreground">{title}</p>
					<p className="text-h1 text-foreground font-bold">{value}</p>
					{trend && (
						<TrendIndicator value={trend.value} label={trend.label} size="sm" />
					)}
				</div>
				<div className="flex flex-col items-end gap-2">
					{Icon && (
						<div
							className={cn("rounded-lg p-2.5", iconColor || "bg-primary/10")}
						>
							<Icon
								className={cn(
									"h-5 w-5",
									iconColor ? "text-current" : "text-primary",
								)}
							/>
						</div>
					)}
					{sparkline}
				</div>
			</div>
		</div>
	);
}

/* ═══════ ChartCard ═══════ */
interface ChartCardProps {
	title: string;
	description?: string;
	actions?: React.ReactNode;
	children: React.ReactNode;
	height?: number;
	className?: string;
}

export function ChartCard({
	title,
	description,
	actions,
	children,
	height = 300,
	className,
}: ChartCardProps) {
	return (
		<div
			className={cn(
				"bg-card border-border p-card-pad rounded-lg border shadow-sm",
				className,
			)}
		>
			<div className="mb-4 flex items-start justify-between">
				<div className="space-y-1">
					<h3 className="text-h4 text-foreground font-semibold">{title}</h3>
					{description && (
						<p className="text-caption text-muted-foreground">{description}</p>
					)}
				</div>
				{actions}
			</div>
			<ResponsiveContainer width="100%" height={height}>
				{children as React.ReactElement}
			</ResponsiveContainer>
		</div>
	);
}

/* ═══════ AnalyticsCard ═══════ */
interface AnalyticsCardProps {
	title: string;
	description?: string;
	actions?: React.ReactNode;
	children: React.ReactNode;
	className?: string;
}

export function AnalyticsCard({
	title,
	description,
	actions,
	children,
	className,
}: AnalyticsCardProps) {
	return (
		<div
			className={cn(
				"bg-card border-border p-card-pad rounded-lg border shadow-sm",
				className,
			)}
		>
			<div className="mb-4 flex items-start justify-between">
				<div className="space-y-1">
					<h3 className="text-h4 text-foreground font-semibold">{title}</h3>
					{description && (
						<p className="text-caption text-muted-foreground">{description}</p>
					)}
				</div>
				{actions}
			</div>
			{children}
		</div>
	);
}

/* ═══════ InsightMetric ═══════ */
interface InsightMetricProps {
	label: string;
	value: string | number;
	status?: "good" | "warning" | "critical" | "neutral";
	description?: string;
}

const statusDot: Record<string, string> = {
	good: "bg-success",
	warning: "bg-warning",
	critical: "bg-danger",
	neutral: "bg-muted-foreground",
};

export function InsightMetric({
	label,
	value,
	status = "neutral",
	description,
}: InsightMetricProps) {
	return (
		<div className="border-border flex items-center justify-between border-b py-3 last:border-0">
			<div className="flex items-center gap-3">
				<span
					className={cn("h-2 w-2 shrink-0 rounded-full", statusDot[status])}
				/>
				<div>
					<p className="text-body text-foreground">{label}</p>
					{description && (
						<p className="text-caption text-muted-foreground">{description}</p>
					)}
				</div>
			</div>
			<span className="text-body text-foreground font-semibold">{value}</span>
		</div>
	);
}
