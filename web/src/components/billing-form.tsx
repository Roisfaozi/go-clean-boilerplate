"use client";

import { Button } from "~/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { toast } from "~/hooks/use-toast";
import { Loader2 } from "lucide-react";
import { useState } from "react";

interface BillingFormProps {
  subscriptionPlan: {
    isPro: boolean;
    isCanceled: boolean;
    stripeSubscriptionId?: string | null;
    stripeCurrentPeriodEnd?: number | null;
    stripeCustomerId?: string | null;
  };
}

export function BillingForm({ subscriptionPlan }: BillingFormProps) {
  const [isLoading, setIsLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setIsLoading(true);

    // TODO: Implement Stripe portal redirection logic
    setTimeout(() => {
      setIsLoading(false);
      toast({
        title: "Billing Portal",
        description: "Redirecting to Stripe portal...",
      });
    }, 1000);
  }

  return (
    <Card>
      <form onSubmit={onSubmit}>
        <CardHeader>
          <CardTitle>Subscription Plan</CardTitle>
          <CardDescription>
            You are currently on the{" "}
            <strong>{subscriptionPlan.isPro ? "Pro" : "Free"}</strong> plan.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">
            {subscriptionPlan.isPro
              ? "Manage your subscription details here."
              : "Upgrade to Pro to unlock all features."}
          </p>
        </CardContent>
        <CardFooter className="flex flex-col items-start space-y-2 text-sm md:flex-row md:justify-between md:space-x-0">
          <Button type="submit" disabled={isLoading}>
            {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {subscriptionPlan.isPro ? "Manage Subscription" : "Upgrade to Pro"}
          </Button>
          {subscriptionPlan.isPro && (
            <p className="rounded-full font-medium">
              {subscriptionPlan.isCanceled
                ? "Your plan will be canceled on "
                : "Your plan renews on "}
              {subscriptionPlan.stripeCurrentPeriodEnd
                ? new Date(
                    subscriptionPlan.stripeCurrentPeriodEnd * 1000
                  ).toLocaleDateString()
                : null}
            </p>
          )}
        </CardFooter>
      </form>
    </Card>
  );
}
