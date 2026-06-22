import { Button } from "~/components/ui/button";
import { Check } from "lucide-react";

const tiers = [
  {
    name: "Starter",
    price: "$49",
    description: "Perfect for solo developers and indie hackers.",
    features: [
      "Single Project License",
      "Next.js 16 Source Code",
      "Tailwind v4 Styling",
      "Basic Auth & Dashboard",
      "Community Support",
    ],
    cta: "Get Started",
    popular: false,
  },
  {
    name: "Team",
    price: "$129",
    description: "For startups and growing teams.",
    features: [
      "Unlimited Projects",
      "Figma Design Files",
      "Advanced Data Grid",
      "AI Chat Integration",
      "Private Repo Access",
      "Priority Support",
    ],
    cta: "Get License",
    popular: true,
  },
  {
    name: "Enterprise",
    price: "$399",
    description: "Full source for SaaS and large scale apps.",
    features: [
      "Extended License (SaaS)",
      "Audit Logs Module",
      "RBAC & Permission Matrix",
      "Legacy Vite/React Version",
      "Direct Founder Support",
      "Custom Contract",
    ],
    cta: "Contact Sales",
    popular: false,
  },
];

export default function Pricing() {
  return (
    <section className="bg-white py-24 dark:bg-slate-950">
      <div className="container px-4 md:px-6">
        <div className="mx-auto mb-16 max-w-3xl text-center">
          <h2 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl dark:text-slate-50">
            Simple, transparent pricing
          </h2>
          <p className="mt-4 text-lg text-slate-500 dark:text-slate-400">
            Choose the license that fits your needs. One-time payment, lifetime
            updates.
          </p>
        </div>

        <div className="mx-auto grid max-w-7xl grid-cols-1 gap-8 md:grid-cols-3">
          {tiers.map((tier) => (
            <div
              key={tier.name}
              className={`relative flex flex-col rounded-3xl border p-8 ${
                tier.popular
                  ? "z-10 scale-105 border-indigo-600 shadow-xl dark:border-indigo-500"
                  : "border-slate-200 bg-slate-50/50 dark:border-slate-800 dark:bg-slate-900/50"
              }`}
            >
              {tier.popular && (
                <div className="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-indigo-600 px-4 py-1 text-sm font-medium text-white">
                  Most Popular
                </div>
              )}

              <div className="mb-8">
                <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-50">
                  {tier.name}
                </h3>
                <div className="mt-4 flex items-baseline text-slate-900 dark:text-slate-50">
                  <span className="text-4xl font-bold tracking-tight">
                    {tier.price}
                  </span>
                  <span className="ml-1 text-xl font-semibold text-slate-500 dark:text-slate-400">
                    /lifetime
                  </span>
                </div>
                <p className="mt-4 text-sm text-slate-500 dark:text-slate-400">
                  {tier.description}
                </p>
              </div>

              <ul className="mb-8 flex-1 space-y-4">
                {tier.features.map((feature) => (
                  <li key={feature} className="flex items-start">
                    <Check className="mr-3 h-5 w-5 shrink-0 text-indigo-500" />
                    <span className="text-sm text-slate-700 dark:text-slate-300">
                      {feature}
                    </span>
                  </li>
                ))}
              </ul>

              <Button
                className={`w-full rounded-full ${
                  tier.popular
                    ? "bg-indigo-600 hover:bg-indigo-700"
                    : "bg-slate-900 dark:bg-slate-50 dark:text-slate-900"
                }`}
                variant={tier.popular ? "default" : "outline"}
              >
                {tier.cta}
              </Button>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
