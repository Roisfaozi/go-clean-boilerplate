import { Link } from "react-router";
import { Check } from "lucide-react";
import { cn } from "@casbin/ui";

const plans = [
  {
    name: "Starter",
    price: "$0",
    cadence: "/ month",
    description: "For small teams getting started with internal tooling.",
    cta: { label: "Start Free", to: "/register" },
    features: [
      "Core workspace",
      "Basic roles",
      "Project management",
      "Standard support",
    ],
    highlight: false,
  },
  {
    name: "Growth",
    price: "$49",
    cadence: "/ user / month",
    description: "For scaling teams that need more access control & automation.",
    cta: { label: "Start Growth", to: "/register" },
    features: [
      "Advanced access control",
      "Audit logs",
      "API keys",
      "Webhook support",
      "Priority support",
    ],
    highlight: true,
  },
  {
    name: "Enterprise",
    price: "Custom",
    cadence: "",
    description: "For organizations with governance and compliance needs.",
    cta: { label: "Contact Sales", to: "/register" },
    features: [
      "SSO & advanced governance",
      "Custom workflows",
      "Dedicated support",
      "Onboarding assistance",
      "Compliance ready",
    ],
    highlight: false,
  },
];

export function LandingPricing() {
  return (
    <section id="pricing" className="border-b border-border bg-background py-20 md:py-24">
      <div className="mx-auto max-w-7xl px-layout">
        <div className="mx-auto mb-14 max-w-2xl text-center">
          <span className="text-caption font-semibold uppercase tracking-wider text-primary">
            Pricing
          </span>
          <h2 className="mt-3 text-3xl font-bold tracking-tight text-foreground md:text-4xl">
            Simple pricing for teams at every stage
          </h2>
          <p className="mt-4 text-body-lg text-muted-foreground">
            Mulai gratis, upgrade saat siap scale. Tidak ada biaya tersembunyi.
          </p>
        </div>

        <div className="grid grid-cols-1 gap-gap md:grid-cols-3">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className={cn(
                "relative flex flex-col rounded-xl border p-card-pad transition-all",
                plan.highlight
                  ? "border-primary bg-card shadow-xl ring-1 ring-primary/20"
                  : "border-border bg-card hover:shadow-md"
              )}
            >
              {plan.highlight && (
                <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-primary px-3 py-1 text-caption font-semibold text-primary-foreground shadow-sm">
                  Most popular
                </span>
              )}

              <div className="mb-5">
                <h3 className="text-h3 font-semibold text-foreground">{plan.name}</h3>
                <p className="mt-1.5 text-body text-muted-foreground">{plan.description}</p>
              </div>

              <div className="mb-6 flex items-baseline gap-1">
                <span className="text-4xl font-bold tracking-tight text-foreground">
                  {plan.price}
                </span>
                {plan.cadence && (
                  <span className="text-body text-muted-foreground">{plan.cadence}</span>
                )}
              </div>

              <Link
                to={plan.cta.to}
                className={cn(
                  "inline-flex h-btn w-full items-center justify-center rounded-md px-5 text-body font-medium shadow-sm transition-colors",
                  plan.highlight
                    ? "bg-primary text-primary-foreground hover:bg-primary-hover"
                    : "border border-border bg-background text-foreground hover:bg-surface-hover"
                )}
              >
                {plan.cta.label}
              </Link>

              <div className="my-6 h-px w-full bg-border" />

              <ul className="space-y-3">
                {plan.features.map((f) => (
                  <li key={f} className="flex items-start gap-2.5 text-body">
                    <Check
                      className={cn(
                        "mt-0.5 h-4 w-4 shrink-0",
                        plan.highlight ? "text-primary" : "text-secondary"
                      )}
                    />
                    <span className="text-foreground">{f}</span>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}