import { Link } from "react-router";
import { PageHeader } from "@/components/layout/page-header";
import { Badge } from "@casbin/ui";
import {
  MousePointerClick,
  Tag,
  FormInput,
  CreditCard,
  MessageSquare,
  Table2,
  Layers,
  Navigation,
  BarChart3,
  Wifi,
  Code2,
} from "lucide-react";

const categories = [
  {
    label: "Buttons",
    description: "All variants, sizes, states, and loading indicators",
    path: "/components/buttons",
    icon: MousePointerClick,
    count: 12,
  },
  {
    label: "Badges",
    description: "Semantic status indicators and labels",
    path: "/components/badges",
    icon: Tag,
    count: 8,
  },
  {
    label: "Form Controls",
    description: "Inputs, pickers, selects, tags, stepper form",
    path: "/components/forms",
    icon: FormInput,
    count: 15,
  },
  {
    label: "Cards",
    description: "NexusCard, StatCard, MetricCard variants",
    path: "/components/cards",
    icon: CreditCard,
    count: 6,
  },
  {
    label: "Feedback",
    description: "Alerts, progress, skeleton, toast, notifications",
    path: "/components/feedback",
    icon: MessageSquare,
    count: 10,
  },
  {
    label: "Data Display",
    description: "DataTable, TreeView, Timeline, CodeBlock",
    path: "/components/data-display",
    icon: Table2,
    count: 8,
  },
  {
    label: "Overlays",
    description: "Dialogs, tooltips, sheets, and popovers",
    path: "/components/overlays",
    icon: Layers,
    count: 6,
  },
  {
    label: "Navigation",
    description: "Sidebar, mega dropdown, breadcrumb, tabs",
    path: "/components/navigation",
    icon: Navigation,
    count: 7,
  },
  {
    label: "Charts",
    description: "Line, Area, Bar, Pie, Heatmap, Sparkline",
    path: "/components/charts",
    icon: BarChart3,
    count: 9,
  },
  {
    label: "Realtime",
    description: "SSE, WebSocket, presence, live activity feed",
    path: "/components/realtime",
    icon: Wifi,
    count: 5,
  },
];

export default function ComponentsIndexPage() {
  return (
    <div className="max-w-5xl space-y-8">
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

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {categories.map((cat) => (
          <Link
            key={cat.path}
            to={cat.path}
            className="group border-border bg-card hover:border-primary/40 relative flex items-start gap-4 rounded-lg border p-6 transition-all duration-200 hover:shadow-md"
          >
            <div className="bg-primary/10 group-hover:bg-primary/20 shrink-0 rounded-lg p-2.5 transition-colors">
              <cat.icon className="text-primary h-5 w-5" />
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <h3 className="text-foreground group-hover:text-primary text-sm font-semibold transition-colors">
                  {cat.label}
                </h3>
                <Badge variant="secondary" className="px-1.5 py-0 text-[10px]">
                  {cat.count}
                </Badge>
              </div>
              <p className="text-muted-foreground mt-1 line-clamp-2 text-xs">
                {cat.description}
              </p>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
