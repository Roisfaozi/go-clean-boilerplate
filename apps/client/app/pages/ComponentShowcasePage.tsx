import { Link } from "react-router";
import { PageHeader } from "@/components/layout/page-header";
import {  Badge  } from "@casbin/ui";
import {
  MousePointerClick, Tag, FormInput, CreditCard, MessageSquare,
  Table2, Layers, Navigation, BarChart3, Wifi, Code2,
} from "lucide-react";

const categories = [
  { label: "Buttons", description: "All variants, sizes, states, and loading indicators", path: "/components/buttons", icon: MousePointerClick, count: 12 },
  { label: "Badges", description: "Semantic status indicators and labels", path: "/components/badges", icon: Tag, count: 8 },
  { label: "Form Controls", description: "Inputs, pickers, selects, tags, stepper form", path: "/components/forms", icon: FormInput, count: 15 },
  { label: "Cards", description: "NexusCard, StatCard, MetricCard variants", path: "/components/cards", icon: CreditCard, count: 6 },
  { label: "Feedback", description: "Alerts, progress, skeleton, toast, notifications", path: "/components/feedback", icon: MessageSquare, count: 10 },
  { label: "Data Display", description: "DataTable, TreeView, Timeline, CodeBlock", path: "/components/data-display", icon: Table2, count: 8 },
  { label: "Overlays", description: "Dialogs, tooltips, sheets, and popovers", path: "/components/overlays", icon: Layers, count: 6 },
  { label: "Navigation", description: "Sidebar, mega dropdown, breadcrumb, tabs", path: "/components/navigation", icon: Navigation, count: 7 },
  { label: "Charts", description: "Line, Area, Bar, Pie, Heatmap, Sparkline", path: "/components/charts", icon: BarChart3, count: 9 },
  { label: "Realtime", description: "SSE, WebSocket, presence, live activity feed", path: "/components/realtime", icon: Wifi, count: 5 },
];

export default function ComponentsIndexPage() {
  return (
    <div className="space-y-8 max-w-5xl">
      <PageHeader
        title="Component Library"
        description="Interactive documentation for all NexusOS UI components. Select a category to explore."
        actions={
          <Badge variant="outline" className="gap-1.5">
            <Code2 className="h-3 w-3" />
            {categories.reduce((sum, c) => sum + c.count, 0)} components
          </Badge>
        }
      />

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {categories.map((cat) => (
          <Link
            key={cat.path}
            to={cat.path}
            className="group relative flex items-start gap-4 p-6 rounded-lg border border-border bg-card hover:border-primary/40 hover:shadow-md transition-all duration-200"
          >
            <div className="rounded-lg bg-primary/10 p-2.5 group-hover:bg-primary/20 transition-colors shrink-0">
              <cat.icon className="h-5 w-5 text-primary" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <h3 className="text-sm font-semibold text-foreground group-hover:text-primary transition-colors">{cat.label}</h3>
                <Badge variant="secondary" className="text-[10px] px-1.5 py-0">{cat.count}</Badge>
              </div>
              <p className="text-xs text-muted-foreground mt-1 line-clamp-2">{cat.description}</p>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
