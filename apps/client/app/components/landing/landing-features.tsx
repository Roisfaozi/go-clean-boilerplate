import {
  LayoutGrid,
  ShieldCheck,
  Activity,
  Plug,
  Palette,
  Lock,
} from "lucide-react";

const features = [
  {
    icon: LayoutGrid,
    title: "Unified Workspace",
    description:
      "Kelola operasi, user, project, dan konfigurasi dari satu platform.",
    tone: "primary",
  },
  {
    icon: ShieldCheck,
    title: "Access Control",
    description:
      "Role, permission, dan organizational scope siap untuk skenario enterprise.",
    tone: "accent",
  },
  {
    icon: Activity,
    title: "Operational Visibility",
    description:
      "Pantau activity, audit trail, insights, dan usage secara real-time.",
    tone: "secondary",
  },
  {
    icon: Plug,
    title: "API-Ready Foundation",
    description:
      "Struktur siap integrasi backend, webhook, upload, dan event-driven workflow.",
    tone: "info",
  },
  {
    icon: Palette,
    title: "Scalable UI System",
    description:
      "Design system konsisten untuk admin, dashboard, dan marketing surfaces.",
    tone: "warning",
  },
  {
    icon: Lock,
    title: "Secure by Design",
    description:
      "Session handling, auth flow, API key pattern, dan governance-friendly architecture.",
    tone: "success",
  },
];

const toneMap: Record<string, string> = {
  primary: "bg-primary/10 text-primary",
  accent: "bg-accent/10 text-accent",
  secondary: "bg-secondary/10 text-secondary",
  info: "bg-info/10 text-info",
  warning: "bg-warning/10 text-warning",
  success: "bg-success/10 text-success",
};

export function LandingFeatures() {
  return (
    <section
      id="features"
      className="border-border bg-surface border-b py-20 md:py-24"
    >
      <div className="px-layout mx-auto max-w-7xl">
        <div className="mx-auto mb-14 max-w-2xl text-center">
          <span className="text-caption text-primary font-semibold tracking-wider uppercase">
            Capabilities
          </span>
          <h2 className="text-foreground mt-3 text-3xl font-bold tracking-tight md:text-4xl">
            Built for teams that need clarity, control, and scale
          </h2>
          <p className="text-body-lg text-muted-foreground mt-4">
            Capability dasar yang mendukung tim engineering, ops, dan product.
          </p>
        </div>

        <div className="gap-gap grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => {
            const Icon = feature.icon;
            return (
              <div
                key={feature.title}
                className="group border-border bg-card p-card-pad hover:border-primary/30 hover:bg-surface-hover rounded-xl border transition-all hover:shadow-md"
              >
                <div
                  className={`mb-4 inline-flex h-11 w-11 items-center justify-center rounded-lg ${toneMap[feature.tone]}`}
                >
                  <Icon className="h-5 w-5" />
                </div>
                <h3 className="text-h4 text-foreground mb-2 font-semibold">
                  {feature.title}
                </h3>
                <p className="text-body text-muted-foreground">
                  {feature.description}
                </p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
