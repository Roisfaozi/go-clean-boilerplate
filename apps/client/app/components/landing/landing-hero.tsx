import { Link } from "react-router";
import { ArrowRight, Sparkles, ShieldCheck, KeySquare, Activity, Users } from "lucide-react";

export function LandingHero() {
  return (
    <section className="relative overflow-hidden border-b border-border">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute -top-24 left-1/4 h-[28rem] w-[28rem] rounded-full bg-primary/15 blur-3xl" />
        <div className="absolute -top-12 right-1/4 h-[28rem] w-[28rem] rounded-full bg-accent/15 blur-3xl" />
      </div>

      <div className="relative mx-auto grid max-w-7xl grid-cols-1 items-center gap-12 px-layout py-20 md:py-28 lg:grid-cols-12 lg:gap-10">
        <div className="lg:col-span-6">
          <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-border bg-surface px-3 py-1.5 text-caption font-medium text-muted-foreground">
            <Sparkles className="h-3.5 w-3.5 text-primary" />
            <span>Enterprise-ready workspace orchestration</span>
          </div>

          <h1 className="text-4xl font-bold tracking-tight text-foreground md:text-5xl lg:text-[56px] lg:leading-[1.05]">
            One platform to manage{" "}
            <span className="bg-gradient-to-r from-primary to-accent bg-clip-text text-transparent">
              operations, access, projects
            </span>{" "}
            and growth
          </h1>

          <p className="mt-5 max-w-xl text-body-lg text-muted-foreground">
            Bangun operasi bisnis, kontrol akses, manajemen project, integrasi API,
            dan observability dalam satu workspace yang konsisten dan scalable.
          </p>

          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <Link
              to="/register"
              className="group inline-flex h-btn items-center justify-center gap-2 rounded-md bg-primary px-6 text-body font-medium text-primary-foreground shadow-md transition-all hover:bg-primary-hover hover:shadow-lg"
            >
              Start Free Trial
              <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
            </Link>
            <a
              href="#pricing"
              className="inline-flex h-btn items-center justify-center rounded-md border border-border bg-background px-6 text-body font-medium text-foreground transition-colors hover:bg-surface-hover"
            >
              Book a Demo
            </a>
          </div>

          <ul className="mt-8 grid grid-cols-2 gap-3 text-caption text-muted-foreground sm:grid-cols-4">
            {[
              { icon: ShieldCheck, label: "No credit card" },
              { icon: KeySquare, label: "SSO ready" },
              { icon: Users, label: "Role-aware" },
              { icon: Activity, label: "Realtime" },
            ].map(({ icon: Icon, label }) => (
              <li key={label} className="inline-flex items-center gap-1.5">
                <Icon className="h-3.5 w-3.5 text-secondary" />
                {label}
              </li>
            ))}
          </ul>
        </div>

        {/* Product mockup */}
        <div className="lg:col-span-6">
          <div className="relative overflow-hidden rounded-xl border border-border bg-card shadow-xl">
            <div className="flex items-center gap-1.5 border-b border-border bg-surface px-4 py-3">
              <span className="h-2.5 w-2.5 rounded-full bg-danger/70" />
              <span className="h-2.5 w-2.5 rounded-full bg-warning/70" />
              <span className="h-2.5 w-2.5 rounded-full bg-success/70" />
              <span className="ml-3 text-caption text-muted-foreground">app.nexusos.io / dashboard</span>
            </div>
            <div className="grid grid-cols-12 gap-3 p-4">
              {/* Side nav */}
              <div className="col-span-3 space-y-1.5 rounded-md bg-surface p-2">
                {["Overview", "Users", "Projects", "Access", "Audit", "Settings"].map((l, i) => (
                  <div
                    key={l}
                    className={`rounded px-2 py-1.5 text-caption ${
                      i === 0 ? "bg-primary/10 text-primary font-medium" : "text-muted-foreground"
                    }`}
                  >
                    {l}
                  </div>
                ))}
              </div>
              {/* Main */}
              <div className="col-span-9 space-y-3">
                <div className="grid grid-cols-3 gap-3">
                  {[
                    { label: "Active users", value: "2,847", track: "bg-primary/15", bar: "bg-primary" },
                    { label: "API calls", value: "1.2M", track: "bg-secondary/15", bar: "bg-secondary" },
                    { label: "Uptime", value: "99.98%", track: "bg-success/15", bar: "bg-success" },
                  ].map((s) => (
                    <div key={s.label} className="rounded-md border border-border bg-background p-3">
                      <div className="text-caption text-muted-foreground">{s.label}</div>
                      <div className="mt-1 text-h3 font-semibold text-foreground">{s.value}</div>
                      <div className={`mt-2 h-1 w-full rounded-full ${s.track}`}>
                        <div className={`h-1 w-2/3 rounded-full ${s.bar}`} />
                      </div>
                    </div>
                  ))}
                </div>
                <div className="rounded-md border border-border bg-background p-3">
                  <div className="mb-2 flex items-center justify-between">
                    <span className="text-caption font-medium text-foreground">Activity</span>
                    <span className="text-caption text-muted-foreground">Last 24h</span>
                  </div>
                  <div className="flex h-20 items-end gap-1.5">
                    {[40, 65, 30, 80, 55, 90, 70, 60, 85, 45, 75, 95].map((h, i) => (
                      <div
                        key={i}
                        className="flex-1 rounded-t bg-gradient-to-t from-primary/30 to-primary"
                        style={{ height: `${h}%` }}
                      />
                    ))}
                  </div>
                </div>
                <div className="space-y-1.5">
                  {[
                    { who: "Sarah", what: "granted admin role", when: "2m" },
                    { who: "Marcus", what: "deployed v2.4.0", when: "8m" },
                    { who: "Aisha", what: "invited 3 members", when: "14m" },
                  ].map((a, i) => (
                    <div key={i} className="flex items-center justify-between rounded border border-border bg-background px-3 py-2 text-caption">
                      <span className="text-foreground">
                        <span className="font-medium">{a.who}</span>{" "}
                        <span className="text-muted-foreground">{a.what}</span>
                      </span>
                      <span className="text-muted-foreground">{a.when} ago</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}