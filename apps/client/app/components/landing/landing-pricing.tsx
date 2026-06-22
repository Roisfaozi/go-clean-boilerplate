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
    description:
      "For scaling teams that need more access control & automation.",
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
    <section
      id="pricing"
      className="border-border bg-background border-b py-20 md:py-24"
    >
      <div className="px-layout mx-auto max-w-7xl">
        <div className="mx-auto mb-14 max-w-2xl text-center">
          <span className="text-caption text-primary font-semibold tracking-wider uppercase">
            Pricing
          </span>
          <h2 className="text-foreground mt-3 text-3xl font-bold tracking-tight md:text-4xl">
            Simple pricing for teams at every stage
          </h2>
          <p className="text-body-lg text-muted-foreground mt-4">
            Mulai gratis, upgrade saat siap scale. Tidak ada biaya tersembunyi.
          </p>
        </div>

        <div className="gap-gap grid grid-cols-1 md:grid-cols-3">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className={cn(
                "p-card-pad relative flex flex-col rounded-xl border transition-all",
                plan.highlight
                  ? "border-primary bg-card ring-primary/20 shadow-xl ring-1"
                  : "border-border bg-card hover:shadow-md",
              )}
            >
              {plan.highlight && (
                <span className="bg-primary text-caption text-primary-foreground absolute -top-3 left-1/2 -translate-x-1/2 rounded-full px-3 py-1 font-semibold shadow-sm">
                  Most popular
                </span>
              )}

              <div className="mb-5">
                <h3 className="text-h3 text-foreground font-semibold">
                  {plan.name}
                </h3>
                <p className="text-body text-muted-foreground mt-1.5">
                  {plan.description}
                </p>
              </div>

              <div className="mb-6 flex items-baseline gap-1">
                <span className="text-foreground text-4xl font-bold tracking-tight">
                  {plan.price}
                </span>
                {plan.cadence && (
                  <span className="text-body text-muted-foreground">
                    {plan.cadence}
                  </span>
                )}
              </div>

              <Link
                to={plan.cta.to}
                className={cn(
                  "h-btn text-body inline-flex w-full items-center justify-center rounded-md px-5 font-medium shadow-sm transition-colors",
                  plan.highlight
                    ? "bg-primary text-primary-foreground hover:bg-primary-hover"
                    : "border-border bg-background text-foreground hover:bg-surface-hover border",
                )}
              >
                {plan.cta.label}
              </Link>

              <div className="bg-border my-6 h-px w-full" />

              <ul className="space-y-3">
                {plan.features.map((f) => (
                  <li key={f} className="text-body flex items-start gap-2.5">
                    <Check
                      className={cn(
                        "mt-0.5 h-4 w-4 shrink-0",
                        plan.highlight ? "text-primary" : "text-secondary",
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
