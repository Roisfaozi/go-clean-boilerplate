import { AlertTriangleIcon } from "lucide-react";
import { Alert, AlertDescription } from "~/components/ui/alert";
import { getCurrentSession } from "~/lib/server/auth/session";
import { getUserSubscriptionPlan, stripe } from "~/lib/server/payment";
import { BillingProvider } from "./_components/billing-context";
import { PlanStatusCard } from "./_components/plan-status-card";
import { PricingComparison } from "./_components/pricing-comparison";

export default async function Billing() {
  const { user } = await getCurrentSession();
  const subscriptionPlan = await getUserSubscriptionPlan(user?.id as string);

  let isCanceled = false;
  if (subscriptionPlan.isPro && subscriptionPlan.stripeSubscriptionId) {
    const stripePlan = await stripe.subscriptions.retrieve(
      subscriptionPlan.stripeSubscriptionId
    );
    isCanceled = stripePlan.cancel_at_period_end;
  }

  const initialPlan = {
    ...subscriptionPlan,
    isCanceled,
  };

  return (
    <BillingProvider initialPlan={initialPlan}>
      <div className="space-y-8">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Billing</h2>
          <p className="text-muted-foreground">
            Manage your subscription, billing details, and invoices.
          </p>
        </div>

        <Alert>
          <div className="flex items-center gap-2">
            <AlertTriangleIcon className="h-5 w-5 shrink-0" />
            <div>
              <AlertDescription>
                <strong>NexusOS Billing</strong> currently runs in test mode.
                Use Stripe test credentials for upgrading.
              </AlertDescription>
            </div>
          </div>
        </Alert>

        <div className="grid gap-8 lg:grid-cols-2">
          <PlanStatusCard />
          <PricingComparison />
        </div>
      </div>
    </BillingProvider>
  );
}
