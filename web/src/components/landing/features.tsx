import { Shield, Zap, Layout, Lock, Database, BarChart } from "lucide-react";
import { BentoCard, BentoGrid } from "~/components/magicui/bento-grid";

const features = [
  {
    name: "Dual-Density Engine",
    description:
      "Switch instantly between Comfort mode for analytics and Compact mode for data-heavy operations.",
    icon: Layout,
    href: "#",
    cta: "Learn more",
    background: <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/10 to-transparent" />,
    className: "lg:col-span-2 lg:row-span-1",
  },
  {
    name: "Enterprise RBAC",
    description: "Granular permission controls with role inheritance.",
    icon: Shield,
    href: "#",
    cta: "Learn more",
    background: <div className="absolute inset-0 bg-gradient-to-tr from-teal-500/10 to-transparent" />,
    className: "lg:col-span-1 lg:row-span-1",
  },
  {
    name: "AI-Native Integration",
    description: "Built-in support for Vercel AI SDK with streaming chat.",
    icon: Zap,
    href: "#",
    cta: "Learn more",
    background: <div className="absolute inset-0 bg-gradient-to-br from-violet-500/10 to-transparent" />,
    className: "lg:col-span-1 lg:row-span-1",
  },
  {
    name: "Audit Logging",
    description: "Comprehensive activity tracking for security compliance.",
    icon: Lock,
    href: "#",
    cta: "Learn more",
    background: <div className="absolute inset-0 bg-gradient-to-bl from-amber-500/10 to-transparent" />,
    className: "lg:col-span-2 lg:row-span-1",
  },
];

export default function Features() {
  return (
    <section className="py-24">
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

        <BentoGrid>
          {features.map((feature) => (
            <BentoCard key={feature.name} {...feature} Icon={feature.icon} />
          ))}
        </BentoGrid>
      </div>
    </section>
  );
}
