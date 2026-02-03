import { Shield, Zap, Layout, Lock, Database, BarChart } from "lucide-react";

const features = [
  {
    name: "Dual-Density Engine",
    description: "Switch instantly between Comfort mode for analytics and Compact mode for data-heavy operations.",
    icon: Layout,
  },
  {
    name: "Enterprise RBAC",
    description: "Granular permission controls with role inheritance and resource-level security.",
    icon: Shield,
  },
  {
    name: "High-Performance Grid",
    description: "Handle 100k+ rows with TanStack Table integration, sticky columns, and server-side operations.",
    icon: Database,
  },
  {
    name: "AI-Native Integration",
    description: "Built-in support for Vercel AI SDK with streaming chat and context-aware assistance.",
    icon: Zap,
  },
  {
    name: "Audit Logging",
    description: "Comprehensive activity tracking for security compliance and troubleshooting.",
    icon: Lock,
  },
  {
    name: "Real-time Analytics",
    description: "Live dashboards powered by Recharts and WebSockets for instant insights.",
    icon: BarChart,
  },
];

export default function Features() {
  return (
    <section className="py-24 bg-slate-50 dark:bg-slate-900/50">
      <div className="container px-4 md:px-6">
        <div className="text-center max-w-3xl mx-auto mb-16">
          <h2 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-slate-50 sm:text-4xl">
            Everything you need to build faster
          </h2>
          <p className="mt-4 text-lg text-slate-500 dark:text-slate-400">
            NexusOS combines the speed of modern SaaS templates with the robustness of enterprise admin systems.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature) => (
            <div key={feature.name} className="relative group p-8 bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm hover:shadow-md transition-all">
              <div className="inline-flex items-center justify-center p-3 rounded-xl bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 mb-6 group-hover:scale-110 transition-transform">
                <feature.icon className="h-6 w-6" />
              </div>
              <h3 className="text-xl font-semibold text-slate-900 dark:text-slate-50 mb-3">
                {feature.name}
              </h3>
              <p className="text-slate-500 dark:text-slate-400 leading-relaxed">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
