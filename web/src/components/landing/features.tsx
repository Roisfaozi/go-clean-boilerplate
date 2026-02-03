import { Shield, Zap, Layout, Lock, Database, BarChart } from "lucide-react";

const features = [
  {
    name: "Dual-Density Engine",
    description:
      "Switch instantly between Comfort mode for analytics and Compact mode for data-heavy operations.",
    icon: Layout,
  },
  {
    name: "Enterprise RBAC",
    description:
      "Granular permission controls with role inheritance and resource-level security.",
    icon: Shield,
  },
  {
    name: "High-Performance Grid",
    description:
      "Handle 100k+ rows with TanStack Table integration, sticky columns, and server-side operations.",
    icon: Database,
  },
  {
    name: "AI-Native Integration",
    description:
      "Built-in support for Vercel AI SDK with streaming chat and context-aware assistance.",
    icon: Zap,
  },
  {
    name: "Audit Logging",
    description:
      "Comprehensive activity tracking for security compliance and troubleshooting.",
    icon: Lock,
  },
  {
    name: "Real-time Analytics",
    description:
      "Live dashboards powered by Recharts and WebSockets for instant insights.",
    icon: BarChart,
  },
];

export default function Features() {
  return (
    <section className="bg-slate-50 py-24 dark:bg-slate-900/50">
      <div className="container px-4 md:px-6">
        <div className="mx-auto mb-16 max-w-3xl text-center">
          <h2 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl dark:text-slate-50">
            Everything you need to build faster
          </h2>
          <p className="mt-4 text-lg text-slate-500 dark:text-slate-400">
            NexusOS combines the speed of modern SaaS templates with the
            robustness of enterprise admin systems.
          </p>
        </div>

        <div className="grid grid-cols-1 gap-8 md:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => (
            <div
              key={feature.name}
              className="group relative rounded-2xl border border-slate-200 bg-white p-8 shadow-sm transition-all hover:shadow-md dark:border-slate-800 dark:bg-slate-900"
            >
              <div className="mb-6 inline-flex items-center justify-center rounded-xl bg-indigo-50 p-3 text-indigo-600 transition-transform group-hover:scale-110 dark:bg-indigo-900/20 dark:text-indigo-400">
                <feature.icon className="h-6 w-6" />
              </div>
              <h3 className="mb-3 text-xl font-semibold text-slate-900 dark:text-slate-50">
                {feature.name}
              </h3>
              <p className="leading-relaxed text-slate-500 dark:text-slate-400">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
