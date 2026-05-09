import { CheckCircle2, ShieldCheck, FolderKanban, ScrollText, Plug } from "lucide-react";

const blocks = [
  {
    icon: ShieldCheck,
    eyebrow: "Identity & Access",
    title: "Manage roles, permissions, and organizational access with precision",
    points: [
      "Role assignment & inheritance",
      "Organization-aware access scope",
      "Guarded actions & route-level control",
      "Audit-ready policy changes",
    ],
    tone: "primary",
    visual: "access",
  },
  {
    icon: FolderKanban,
    eyebrow: "Project & Workspace",
    title: "Coordinate teams, projects, and execution across workspaces",
    points: [
      "Project overview & ownership",
      "Membership & role mapping",
      "Environment-aware structure",
      "Operational drilldown",
    ],
    tone: "secondary",
    visual: "projects",
  },
  {
    icon: ScrollText,
    eyebrow: "Audit & Governance",
    title: "Track actions, monitor health, and stay audit-ready",
    points: [
      "Detailed audit logs",
      "Live activity streams",
      "Insight summaries",
      "System health & history",
    ],
    tone: "accent",
    visual: "audit",
  },
  {
    icon: Plug,
    eyebrow: "Integrations & Realtime",
    title: "Connect APIs, handle uploads, and enable realtime workflows",
    points: [
      "Webhook-ready endpoints",
      "API key management",
      "Resumable chunked uploads",
      "WebSocket / SSE event streams",
    ],
    tone: "info",
    visual: "integrations",
  },
];

const toneText: Record<string, string> = {
  primary: "text-primary",
  secondary: "text-secondary",
  accent: "text-accent",
  info: "text-info",
};
const toneBg: Record<string, string> = {
  primary: "bg-primary/10",
  secondary: "bg-secondary/10",
  accent: "bg-accent/10",
  info: "bg-info/10",
};

function VisualMock({ kind, tone }: { kind: string; tone: string }) {
  const accent = toneText[tone];
  if (kind === "access") {
    return (
      <div className="rounded-xl border border-border bg-card p-card-pad shadow-lg">
        <div className="mb-3 flex items-center justify-between">
          <span className="text-caption font-medium text-foreground">Role Matrix</span>
          <span className="text-caption text-muted-foreground">3 roles</span>
        </div>
        <div className="overflow-hidden rounded-md border border-border">
          <div className="grid grid-cols-4 bg-surface text-caption font-medium text-muted-foreground">
            {["Resource", "Admin", "Editor", "Viewer"].map((h) => (
              <div key={h} className="px-3 py-2">{h}</div>
            ))}
          </div>
          {["Users", "Projects", "Billing", "Audit"].map((r) => (
            <div key={r} className="grid grid-cols-4 border-t border-border text-caption">
              <div className="px-3 py-2 font-medium text-foreground">{r}</div>
              {[true, r !== "Billing", r === "Users" || r === "Projects"].map((on, i) => (
                <div key={i} className="px-3 py-2">
                  <span className={`inline-block h-2 w-2 rounded-full ${on ? "bg-success" : "bg-border-strong"}`} />
                </div>
              ))}
            </div>
          ))}
        </div>
      </div>
    );
  }
  if (kind === "projects") {
    return (
      <div className="rounded-xl border border-border bg-card p-card-pad shadow-lg">
        <div className="mb-3 text-caption font-medium text-foreground">Active projects</div>
        <div className="space-y-2">
          {[
            { n: "Atlas Migration", p: 78, m: 12 },
            { n: "Helio Onboarding", p: 45, m: 7 },
            { n: "Orion Insights v2", p: 92, m: 4 },
          ].map((p) => (
            <div key={p.n} className="rounded-md border border-border bg-background p-3">
              <div className="mb-1.5 flex items-center justify-between text-caption">
                <span className="font-medium text-foreground">{p.n}</span>
                <span className="text-muted-foreground">{p.m} members</span>
              </div>
              <div className="h-1.5 w-full overflow-hidden rounded-full bg-surface">
                <div className={`h-full bg-secondary`} style={{ width: `${p.p}%` }} />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }
  if (kind === "audit") {
    return (
      <div className="rounded-xl border border-border bg-card p-card-pad shadow-lg">
        <div className="mb-3 flex items-center justify-between text-caption">
          <span className="font-medium text-foreground">Audit timeline</span>
          <span className="text-muted-foreground">Today</span>
        </div>
        <ol className="space-y-3">
          {[
            { t: "10:42", who: "sarah@", what: "updated billing role" },
            { t: "10:18", who: "marcus@", what: "rotated API key" },
            { t: "09:55", who: "system", what: "health check passed" },
            { t: "09:30", who: "aisha@", what: "invited 2 members" },
          ].map((e) => (
            <li key={e.t} className="flex items-start gap-3 text-caption">
              <span className={`mt-1 h-2 w-2 shrink-0 rounded-full ${toneBg[tone]} ${accent}`} />
              <div className="flex-1">
                <span className="text-muted-foreground">{e.t}</span>{" "}
                <span className="font-medium text-foreground">{e.who}</span>{" "}
                <span className="text-muted-foreground">{e.what}</span>
              </div>
            </li>
          ))}
        </ol>
      </div>
    );
  }
  // integrations
  return (
    <div className="rounded-xl border border-border bg-card p-card-pad shadow-lg">
      <div className="mb-3 text-caption font-medium text-foreground">Connected services</div>
      <div className="grid grid-cols-2 gap-2">
        {[
          { n: "Webhooks", s: "12 active" },
          { n: "Uploads", s: "Resumable" },
          { n: "API Keys", s: "4 keys" },
          { n: "Realtime", s: "WS + SSE" },
        ].map((i) => (
          <div key={i.n} className="rounded-md border border-border bg-background p-3">
            <div className="text-caption font-medium text-foreground">{i.n}</div>
            <div className="text-caption text-muted-foreground">{i.s}</div>
            <div className="mt-2 inline-flex items-center gap-1 text-[11px] text-success">
              <span className="h-1.5 w-1.5 rounded-full bg-success" /> Healthy
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export function LandingFeatureSplit() {
  return (
    <section id="modules" className="border-b border-border bg-background py-20 md:py-24">
      <div className="mx-auto max-w-7xl px-layout">
        <div className="mx-auto mb-16 max-w-2xl text-center">
          <span className="text-caption font-semibold uppercase tracking-wider text-primary">
            Modules
          </span>
          <h2 className="mt-3 text-3xl font-bold tracking-tight text-foreground md:text-4xl">
            Purpose-built modules for modern internal platforms
          </h2>
          <p className="mt-4 text-body-lg text-muted-foreground">
            Setiap modul dirancang untuk kebutuhan operasional yang nyata.
          </p>
        </div>

        <div className="space-y-20 md:space-y-28">
          {blocks.map((b, idx) => {
            const Icon = b.icon;
            const reverse = idx % 2 === 1;
            return (
              <div
                key={b.title}
                className="grid grid-cols-1 items-center gap-10 lg:grid-cols-2 lg:gap-16"
              >
                <div className={reverse ? "lg:order-2" : ""}>
                  <div className={`mb-4 inline-flex h-11 w-11 items-center justify-center rounded-lg ${toneBg[b.tone]} ${toneText[b.tone]}`}>
                    <Icon className="h-5 w-5" />
                  </div>
                  <span className={`text-caption font-semibold uppercase tracking-wider ${toneText[b.tone]}`}>
                    {b.eyebrow}
                  </span>
                  <h3 className="mt-2 text-2xl font-bold tracking-tight text-foreground md:text-3xl">
                    {b.title}
                  </h3>
                  <ul className="mt-6 space-y-3">
                    {b.points.map((p) => (
                      <li key={p} className="flex items-start gap-2.5 text-body text-foreground">
                        <CheckCircle2 className={`mt-0.5 h-4 w-4 shrink-0 ${toneText[b.tone]}`} />
                        <span className="text-muted-foreground">{p}</span>
                      </li>
                    ))}
                  </ul>
                </div>
                <div className={reverse ? "lg:order-1" : ""}>
                  <VisualMock kind={b.visual} tone={b.tone} />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}