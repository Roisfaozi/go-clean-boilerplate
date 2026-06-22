import { PageHeader } from "@/components/layout/page-header";
import {
  NexusLineChart,
  NexusAreaChart,
  NexusBarChart,
  NexusPieChart,
  NexusDonutChart,
  NexusHeatmap,
  Sparkline,
} from "@/components/charts/charts";

const chartData = [
  { month: "Jan", users: 120, revenue: 4200, sessions: 800 },
  { month: "Feb", users: 180, revenue: 5100, sessions: 1200 },
  { month: "Mar", users: 250, revenue: 6800, sessions: 1800 },
  { month: "Apr", users: 310, revenue: 8200, sessions: 2400 },
  { month: "May", users: 420, revenue: 9100, sessions: 3100 },
  { month: "Jun", users: 530, revenue: 11200, sessions: 3800 },
];

const pieData = [
  { name: "Admin", value: 12 },
  { name: "Editor", value: 35 },
  { name: "Viewer", value: 53 },
];

const heatmapData = (() => {
  const days = ["Mon", "Tue", "Wed", "Thu", "Fri"];
  const hours = ["9am", "10am", "11am", "12pm", "1pm", "2pm", "3pm", "4pm"];
  return days.flatMap((y) =>
    hours.map((x) => ({ x, y, value: Math.floor(Math.random() * 50) })),
  );
})();

export default function ShowcaseCharts() {
  return (
    <div className="max-w-6xl space-y-10">
      <PageHeader
        title="Charts"
        description="Line, Area, Bar, Pie, Donut, Heatmap, and Sparkline charts."
      />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <NexusLineChart
          title="User Growth"
          description="Monthly active users"
          data={chartData}
          xKey="month"
          lines={[{ dataKey: "users", name: "Users" }]}
        />
        <NexusAreaChart
          title="Revenue"
          description="Monthly revenue trend"
          data={chartData}
          xKey="month"
          areas={[{ dataKey: "revenue", name: "Revenue" }]}
        />
        <NexusBarChart
          title="Sessions vs Users"
          description="Monthly comparison"
          data={chartData}
          xKey="month"
          bars={[
            { dataKey: "sessions", name: "Sessions" },
            { dataKey: "users", name: "Users" },
          ]}
        />
        <NexusPieChart title="Role Distribution" data={pieData} />
        <NexusDonutChart title="Role Distribution (Donut)" data={pieData} />
        <NexusHeatmap
          title="Activity Heatmap"
          description="Commits by day and hour"
          data={heatmapData}
          xLabels={["9am", "10am", "11am", "12pm", "1pm", "2pm", "3pm", "4pm"]}
          yLabels={["Mon", "Tue", "Wed", "Thu", "Fri"]}
        />
      </div>

      <section className="space-y-4">
        <h2 className="text-h2 text-foreground">Sparklines</h2>
        <div className="flex items-center gap-8">
          <div className="flex items-center gap-3">
            <span className="text-body text-foreground">Revenue</span>
            <Sparkline
              data={[10, 25, 18, 30, 22, 40, 35]}
              color="hsl(160, 84%, 39%)"
            />
          </div>
          <div className="flex items-center gap-3">
            <span className="text-body text-foreground">Errors</span>
            <Sparkline
              data={[5, 12, 8, 15, 20, 10, 3]}
              color="hsl(0, 72%, 51%)"
            />
          </div>
          <div className="flex items-center gap-3">
            <span className="text-body text-foreground">Users</span>
            <Sparkline data={[20, 30, 45, 55, 60, 72, 80]} />
          </div>
        </div>
      </section>
    </div>
  );
}
