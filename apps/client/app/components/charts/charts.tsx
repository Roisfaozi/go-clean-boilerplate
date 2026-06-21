import { cn } from "@casbin/ui";
import {
  LineChart as ReLineChart,
  Line,
  AreaChart as ReAreaChart,
  Area,
  BarChart as ReBarChart,
  Bar,
  PieChart as RePieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";

const CHART_COLORS = [
  "hsl(239, 84%, 67%)", // primary
  "hsl(168, 76%, 42%)", // secondary
  "hsl(258, 90%, 66%)", // accent
  "hsl(217, 91%, 60%)", // info
  "hsl(160, 84%, 39%)", // success
  "hsl(38, 92%, 50%)", // warning
  "hsl(0, 72%, 51%)", // danger
];

interface ChartCardProps {
  title?: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}

function ChartCard({
  title,
  description,
  children,
  className,
}: ChartCardProps) {
  return (
    <div
      className={cn(
        "bg-card border-border p-card-pad rounded-lg border shadow-sm",
        className,
      )}
    >
      {(title || description) && (
        <div className="mb-4 space-y-1">
          {title && (
            <h3 className="text-h4 text-foreground font-semibold">{title}</h3>
          )}
          {description && (
            <p className="text-caption text-muted-foreground">{description}</p>
          )}
        </div>
      )}
      {children}
    </div>
  );
}

/* ── LineChart ── */
interface NexusLineChartProps {
  data: Record<string, unknown>[];
  xKey: string;
  lines: { dataKey: string; color?: string; name?: string }[];
  height?: number;
  title?: string;
  description?: string;
  className?: string;
}

export function NexusLineChart({
  data,
  xKey,
  lines,
  height = 300,
  title,
  description,
  className,
}: NexusLineChartProps) {
  return (
    <ChartCard title={title} description={description} className={className}>
      <ResponsiveContainer width="100%" height={height}>
        <ReLineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
          <XAxis
            dataKey={xKey}
            tick={{ fontSize: 12 }}
            stroke="hsl(var(--muted-foreground))"
          />
          <YAxis
            tick={{ fontSize: 12 }}
            stroke="hsl(var(--muted-foreground))"
          />
          <Tooltip
            contentStyle={{
              background: "hsl(var(--popover))",
              border: "1px solid hsl(var(--border))",
              borderRadius: 8,
              fontSize: 13,
            }}
          />
          <Legend />
          {lines.map((line, i) => (
            <Line
              key={line.dataKey}
              type="monotone"
              dataKey={line.dataKey}
              name={line.name || line.dataKey}
              stroke={line.color || CHART_COLORS[i % CHART_COLORS.length]}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4 }}
            />
          ))}
        </ReLineChart>
      </ResponsiveContainer>
    </ChartCard>
  );
}

/* ── AreaChart ── */
interface NexusAreaChartProps {
  data: Record<string, unknown>[];
  xKey: string;
  areas: { dataKey: string; color?: string; name?: string }[];
  height?: number;
  title?: string;
  description?: string;
  className?: string;
}

export function NexusAreaChart({
  data,
  xKey,
  areas,
  height = 300,
  title,
  description,
  className,
}: NexusAreaChartProps) {
  return (
    <ChartCard title={title} description={description} className={className}>
      <ResponsiveContainer width="100%" height={height}>
        <ReAreaChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
          <XAxis
            dataKey={xKey}
            tick={{ fontSize: 12 }}
            stroke="hsl(var(--muted-foreground))"
          />
          <YAxis
            tick={{ fontSize: 12 }}
            stroke="hsl(var(--muted-foreground))"
          />
          <Tooltip
            contentStyle={{
              background: "hsl(var(--popover))",
              border: "1px solid hsl(var(--border))",
              borderRadius: 8,
              fontSize: 13,
            }}
          />
          <Legend />
          {areas.map((area, i) => {
            const color = area.color || CHART_COLORS[i % CHART_COLORS.length];
            return (
              <Area
                key={area.dataKey}
                type="monotone"
                dataKey={area.dataKey}
                name={area.name || area.dataKey}
                stroke={color}
                fill={color}
                fillOpacity={0.1}
                strokeWidth={2}
              />
            );
          })}
        </ReAreaChart>
      </ResponsiveContainer>
    </ChartCard>
  );
}

/* ── BarChart ── */
interface NexusBarChartProps {
  data: Record<string, unknown>[];
  xKey: string;
  bars: { dataKey: string; color?: string; name?: string }[];
  height?: number;
  title?: string;
  description?: string;
  stacked?: boolean;
  className?: string;
}

export function NexusBarChart({
  data,
  xKey,
  bars,
  height = 300,
  title,
  description,
  stacked,
  className,
}: NexusBarChartProps) {
  return (
    <ChartCard title={title} description={description} className={className}>
      <ResponsiveContainer width="100%" height={height}>
        <ReBarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
          <XAxis
            dataKey={xKey}
            tick={{ fontSize: 12 }}
            stroke="hsl(var(--muted-foreground))"
          />
          <YAxis
            tick={{ fontSize: 12 }}
            stroke="hsl(var(--muted-foreground))"
          />
          <Tooltip
            contentStyle={{
              background: "hsl(var(--popover))",
              border: "1px solid hsl(var(--border))",
              borderRadius: 8,
              fontSize: 13,
            }}
          />
          <Legend />
          {bars.map((bar, i) => (
            <Bar
              key={bar.dataKey}
              dataKey={bar.dataKey}
              name={bar.name || bar.dataKey}
              fill={bar.color || CHART_COLORS[i % CHART_COLORS.length]}
              stackId={stacked ? "stack" : undefined}
              radius={[4, 4, 0, 0]}
            />
          ))}
        </ReBarChart>
      </ResponsiveContainer>
    </ChartCard>
  );
}

/* ── PieChart / DonutChart ── */
interface NexusPieChartProps {
  data: { name: string; value: number; color?: string }[];
  height?: number;
  title?: string;
  description?: string;
  donut?: boolean;
  className?: string;
}

export function NexusPieChart({
  data,
  height = 300,
  title,
  description,
  donut,
  className,
}: NexusPieChartProps) {
  return (
    <ChartCard title={title} description={description} className={className}>
      <ResponsiveContainer width="100%" height={height}>
        <RePieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={donut ? "55%" : 0}
            outerRadius="80%"
            paddingAngle={2}
            dataKey="value"
            label={({ name, percent }) =>
              `${name} ${(percent * 100).toFixed(0)}%`
            }
            labelLine={false}
          >
            {data.map((entry, i) => (
              <Cell
                key={i}
                fill={entry.color || CHART_COLORS[i % CHART_COLORS.length]}
              />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              background: "hsl(var(--popover))",
              border: "1px solid hsl(var(--border))",
              borderRadius: 8,
              fontSize: 13,
            }}
          />
          <Legend />
        </RePieChart>
      </ResponsiveContainer>
    </ChartCard>
  );
}

export function NexusDonutChart(props: Omit<NexusPieChartProps, "donut">) {
  return <NexusPieChart {...props} donut />;
}

/* ── Heatmap ── */
interface HeatmapCell {
  x: string;
  y: string;
  value: number;
}

interface NexusHeatmapProps {
  data: HeatmapCell[];
  xLabels: string[];
  yLabels: string[];
  title?: string;
  description?: string;
  className?: string;
}

export function NexusHeatmap({
  data,
  xLabels,
  yLabels,
  title,
  description,
  className,
}: NexusHeatmapProps) {
  const maxVal = Math.max(...data.map((d) => d.value), 1);

  return (
    <ChartCard title={title} description={description} className={className}>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr>
              <th className="w-20" />
              {xLabels.map((x) => (
                <th
                  key={x}
                  className="text-caption text-muted-foreground px-1 py-2 text-center font-normal"
                >
                  {x}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {yLabels.map((y) => (
              <tr key={y}>
                <td className="text-caption text-muted-foreground py-1 pr-3 text-right">
                  {y}
                </td>
                {xLabels.map((x) => {
                  const cell = data.find((d) => d.x === x && d.y === y);
                  const intensity = cell ? cell.value / maxVal : 0;
                  return (
                    <td key={x} className="p-0.5">
                      <div
                        className="text-caption flex aspect-square items-center justify-center rounded-sm"
                        style={{
                          backgroundColor: `hsl(239, 84%, 67%, ${intensity * 0.8 + 0.05})`,
                          color:
                            intensity > 0.5
                              ? "white"
                              : "hsl(var(--muted-foreground))",
                        }}
                        title={`${x}, ${y}: ${cell?.value ?? 0}`}
                      >
                        {cell?.value ?? 0}
                      </div>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </ChartCard>
  );
}

/* ── Sparkline ── */
interface SparklineProps {
  data: number[];
  color?: string;
  width?: number;
  height?: number;
  className?: string;
}

export function Sparkline({
  data,
  color = "hsl(239, 84%, 67%)",
  width = 120,
  height = 32,
  className,
}: SparklineProps) {
  if (data.length < 2) return null;
  const max = Math.max(...data);
  const min = Math.min(...data);
  const range = max - min || 1;
  const points = data
    .map(
      (v, i) =>
        `${(i / (data.length - 1)) * width},${height - ((v - min) / range) * height}`,
    )
    .join(" ");

  return (
    <svg
      width={width}
      height={height}
      className={cn("inline-block", className)}
    >
      <polyline fill="none" stroke={color} strokeWidth={1.5} points={points} />
    </svg>
  );
}
