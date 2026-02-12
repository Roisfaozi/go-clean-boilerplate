"use client";

import { useBilling } from "./billing-context";
import { Button } from "~/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "~/components/ui/card";
import { Icon } from "~/components/shared/icon";

export function PlanStatusCard() {
  const { subscriptionPlan, isLoading, handlePortalRedirect } = useBilling();

  return (
    <Card>
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
            ? "Manage your subscription details and invoices."
            : "Upgrade to Pro to unlock unlimited projects and advanced RBAC features."}
        </p>
      </CardContent>
      <CardFooter className="bg-muted/20 flex flex-col items-start space-y-4 border-t px-6 py-4 md:flex-row md:justify-between md:space-y-0">
        <Button onClick={handlePortalRedirect} disabled={isLoading}>
          {isLoading ? (
            <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Icon name="CreditCard" className="mr-2 h-4 w-4" />
          )}
          {subscriptionPlan.isPro ? "Manage Billing" : "Upgrade to Pro"}
        </Button>

        {subscriptionPlan.isPro && (
          <div className="text-muted-foreground text-sm font-medium">
            {subscriptionPlan.isCanceled
              ? "Plan will be canceled on "
              : "Plan renews on "}
            {subscriptionPlan.stripeCurrentPeriodEnd
              ? new Date(
                  subscriptionPlan.stripeCurrentPeriodEnd * 1000
                ).toLocaleDateString()
              : null}
          </div>
        )}
      </CardFooter>
    </Card>
  );
}
