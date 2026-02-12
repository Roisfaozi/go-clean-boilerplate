"use client";

import { Check } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";

const features = [
  {
    name: "Unlimited Projects",
    free: false,
    pro: true,
  },
  {
    name: "Advanced RBAC (Casbin)",
    free: "Basic",
    pro: "Granular",
  },
  {
    name: "Audit Logs Retention",
    free: "7 Days",
    pro: "Unlimited",
  },
  {
    name: "Priority Support",
    free: false,
    pro: true,
  },
];

export function PricingComparison() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Plan Comparison</CardTitle>
        <CardDescription>
          Detailed breakdown of features included in each plan.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid gap-4">
          {features.map((feature) => (
            <div
              key={feature.name}
              className="flex items-center justify-between border-b pb-4 last:border-0 last:pb-0"
            >
              <div className="text-sm font-medium">{feature.name}</div>
              <div className="flex gap-8 text-sm">
                <div className="w-20 text-center">
                  {typeof feature.free === "boolean" ? (
                    feature.free ? (
                      <Check className="mx-auto h-4 w-4 text-emerald-500" />
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )
                  ) : (
                    <span className="text-muted-foreground">{feature.free}</span>
                  )}
                </div>
                <div className="w-20 text-center font-semibold">
                  {typeof feature.pro === "boolean" ? (
                    feature.pro ? (
                      <Check className="mx-auto h-4 w-4 text-emerald-500" />
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )
                  ) : (
                    <span>{feature.pro}</span>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
