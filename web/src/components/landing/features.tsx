import { Shield, Zap, Layout, Lock, Database, BarChart, Server, Globe } from "lucide-react";
import { BentoCard, BentoGrid } from "~/components/magicui/bento-grid";
import { cn } from "~/lib/utils";

const features = [
  {
    name: "Adaptive Density Engine",
    description: "Switch between Comfort (SaaS) and Compact (Enterprise) modes instantly.",
    icon: Layout,
    href: "/dashboard",
    cta: "Try it out",
    background: (
      <div className="absolute inset-0 bg-linear-to-br from-primary/10 to-transparent transition-opacity group-hover:opacity-20" />
    ),
    className: "lg:col-span-2 lg:row-span-1",
  },
  {
    name: "Casbin RBAC",
    description: "Enterprise-grade authorization with role inheritance and matrix management.",
    icon: Shield,
    href: "/dashboard/access",
    cta: "Manage access",
    background: (
      <div className="absolute inset-0 bg-linear-to-tr from-emerald-500/10 to-transparent transition-opacity group-hover:opacity-20" />
    ),
    className: "lg:col-span-1 lg:row-span-1",
  },
  {
    name: "Multi-tenant Ready",
    description: "Built-in organization switching and strict data isolation for every tenant.",
    icon: Globe,
    href: "/dashboard/organization/settings",
    cta: "Configure org",
    background: (
      <div className="absolute inset-0 bg-linear-to-br from-indigo-500/10 to-transparent transition-opacity group-hover:opacity-20" />
    ),
    className: "lg:col-span-1 lg:row-span-1",
  },
  {
    name: "Real-time Distributed WS",
    description: "WebSocket scaling with Redis Pub/Sub and presence tracking out of the box.",
    icon: Zap,
    href: "/dashboard",
    cta: "See live activity",
    background: (
      <div className="absolute inset-0 bg-linear-to-bl from-amber-500/10 to-transparent transition-opacity group-hover:opacity-20" />
    ),
    className: "lg:col-span-2 lg:row-span-1",
  },
];

export default function Features() {
  return (
    <section className="bg-background/50 py-24 backdrop-blur-sm">
      <div className="container px-4 md:px-6">
        <div className="mx-auto mb-16 max-w-3xl text-center">
          <div className="text-primary mb-4 text-sm font-bold tracking-widest uppercase">
            Core Capabilities
          </div>
          <h2 className="mb-4 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl dark:text-slate-50">
            Enterprise foundation, SaaS speed.
          </h2>
          <p className="text-muted-foreground text-lg">
            NexusOS provides the heavy lifting so you can focus on building your core features. 
            All powered by Go and Next.js.
          </p>
        </div>

        <BentoGrid className="lg:grid-rows-2">
          {features.map((feature) => (
            <BentoCard key={feature.name} {...feature} Icon={feature.icon} />
          ))}
        </BentoGrid>
      </div>
    </section>
  );
}
