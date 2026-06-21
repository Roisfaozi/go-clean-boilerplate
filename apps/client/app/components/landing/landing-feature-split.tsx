import {
  CheckCircle2,
  ShieldCheck,
  FolderKanban,
  ScrollText,
  Plug,
} from "lucide-react";

const blocks = [
  {
    icon: ShieldCheck,
    eyebrow: "Identity & Access",
    title:
      "Manage roles, permissions, and organizational access with precision",
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
      <div className="border-border bg-card p-card-pad rounded-xl border shadow-lg">
        <div className="mb-3 flex items-center justify-between">
          <span className="text-caption text-foreground font-medium">
            Role Matrix
          </span>
          <span className="text-caption text-muted-foreground">3 roles</span>
        </div>
        <div className="border-border overflow-hidden rounded-md border">
          <div className="bg-surface text-caption text-muted-foreground grid grid-cols-4 font-medium">
            {["Resource", "Admin", "Editor", "Viewer"].map((h) => (
              <div key={h} className="px-3 py-2">
                {h}
              </div>
            ))}
          </div>
          {["Users", "Projects", "Billing", "Audit"].map((r) => (
            <div
              key={r}
              className="border-border text-caption grid grid-cols-4 border-t"
            >
              <div className="text-foreground px-3 py-2 font-medium">{r}</div>
              {[true, r !== "Billing", r === "Users" || r === "Projects"].map(
                (on, i) => (
                  <div key={i} className="px-3 py-2">
                    <span
                      className={`inline-block h-2 w-2 rounded-full ${on ? "bg-success" : "bg-border-strong"}`}
                    />
                  </div>
                ),
              )}
            </div>
          ))}
        </div>
      </div>
    );
  }
  if (kind === "projects") {
    return (
      <div className="border-border bg-card p-card-pad rounded-xl border shadow-lg">
        <div className="text-caption text-foreground mb-3 font-medium">
          Active projects
        </div>
        <div className="space-y-2">
          {[
            { n: "Atlas Migration", p: 78, m: 12 },
            { n: "Helio Onboarding", p: 45, m: 7 },
            { n: "Orion Insights v2", p: 92, m: 4 },
          ].map((p) => (
            <div
              key={p.n}
              className="border-border bg-background rounded-md border p-3"
            >
              <div className="text-caption mb-1.5 flex items-center justify-between">
                <span className="text-foreground font-medium">{p.n}</span>
                <span className="text-muted-foreground">{p.m} members</span>
              </div>
              <div className="bg-surface h-1.5 w-full overflow-hidden rounded-full">
                <div
                  className={`bg-secondary h-full`}
                  style={{ width: `${p.p}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }
  if (kind === "audit") {
    return (
      <div className="border-border bg-card p-card-pad rounded-xl border shadow-lg">
        <div className="text-caption mb-3 flex items-center justify-between">
          <span className="text-foreground font-medium">Audit timeline</span>
          <span className="text-muted-foreground">Today</span>
        </div>
        <ol className="space-y-3">
          {[
            { t: "10:42", who: "sarah@", what: "updated billing role" },
            { t: "10:18", who: "marcus@", what: "rotated API key" },
            { t: "09:55", who: "system", what: "health check passed" },
            { t: "09:30", who: "aisha@", what: "invited 2 members" },
          ].map((e) => (
            <li key={e.t} className="text-caption flex items-start gap-3">
              <span
                className={`mt-1 h-2 w-2 shrink-0 rounded-full ${toneBg[tone]} ${accent}`}
              />
              <div className="flex-1">
                <span className="text-muted-foreground">{e.t}</span>{" "}
                <span className="text-foreground font-medium">{e.who}</span>{" "}
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
    <div className="border-border bg-card p-card-pad rounded-xl border shadow-lg">
      <div className="text-caption text-foreground mb-3 font-medium">
        Connected services
      </div>
      <div className="grid grid-cols-2 gap-2">
        {[
          { n: "Webhooks", s: "12 active" },
          { n: "Uploads", s: "Resumable" },
          { n: "API Keys", s: "4 keys" },
          { n: "Realtime", s: "WS + SSE" },
        ].map((i) => (
          <div
            key={i.n}
            className="border-border bg-background rounded-md border p-3"
          >
            <div className="text-caption text-foreground font-medium">
              {i.n}
            </div>
            <div className="text-caption text-muted-foreground">{i.s}</div>
            <div className="text-success mt-2 inline-flex items-center gap-1 text-[11px]">
              <span className="bg-success h-1.5 w-1.5 rounded-full" /> Healthy
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export function LandingFeatureSplit() {
  return (
    <section
      id="modules"
      className="border-border bg-background border-b py-20 md:py-24"
    >
      <div className="px-layout mx-auto max-w-7xl">
        <div className="mx-auto mb-16 max-w-2xl text-center">
          <span className="text-caption text-primary font-semibold tracking-wider uppercase">
            Modules
          </span>
          <h2 className="text-foreground mt-3 text-3xl font-bold tracking-tight md:text-4xl">
            Purpose-built modules for modern internal platforms
          </h2>
          <p className="text-body-lg text-muted-foreground mt-4">
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
                  <div
                    className={`mb-4 inline-flex h-11 w-11 items-center justify-center rounded-lg ${toneBg[b.tone]} ${toneText[b.tone]}`}
                  >
                    <Icon className="h-5 w-5" />
                  </div>
                  <span
                    className={`text-caption font-semibold tracking-wider uppercase ${toneText[b.tone]}`}
                  >
                    {b.eyebrow}
                  </span>
                  <h3 className="text-foreground mt-2 text-2xl font-bold tracking-tight md:text-3xl">
                    {b.title}
                  </h3>
                  <ul className="mt-6 space-y-3">
                    {b.points.map((p) => (
                      <li
                        key={p}
                        className="text-body text-foreground flex items-start gap-2.5"
                      >
                        <CheckCircle2
                          className={`mt-0.5 h-4 w-4 shrink-0 ${toneText[b.tone]}`}
                        />
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
