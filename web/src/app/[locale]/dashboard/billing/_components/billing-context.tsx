"use client";

import {
  createContext,
  useContext,
  useState,
  useCallback,
  ReactNode,
} from "react";
import { toast } from "sonner";

export interface SubscriptionPlan {
  isPro: boolean;
  isCanceled: boolean;
  stripeSubscriptionId?: string | null;
  stripeCurrentPeriodEnd?: number | null;
  stripeCustomerId?: string | null;
}

interface BillingContextType {
  subscriptionPlan: SubscriptionPlan;
  isLoading: boolean;
  handlePortalRedirect: () => Promise<void>;
}

const BillingContext = createContext<BillingContextType | undefined>(undefined);

export function BillingProvider({
  initialPlan,
  children,
}: {
  initialPlan: SubscriptionPlan;
  children: ReactNode;
}) {
  const [isLoading, setIsLoading] = useState(false);
  const [subscriptionPlan] = useState<SubscriptionPlan>(initialPlan);

  const handlePortalRedirect = useCallback(async () => {
    setIsLoading(true);
    try {
      // In a real app, this would call a server action or API to get the Stripe URL
      toast.info("Redirecting to Stripe billing portal...");

      // Simulation
      await new Promise((resolve) => setTimeout(resolve, 1500));

      // window.location.href = stripeUrl;
    } catch (error) {
      toast.error("Failed to open billing portal");
    } finally {
      setIsLoading(false);
    }
  }, []);

  return (
    <BillingContext.Provider
      value={{
        subscriptionPlan,
        isLoading,
        handlePortalRedirect,
      }}
    >
      {children}
    </BillingContext.Provider>
  );
}

export function useBilling() {
  const context = useContext(BillingContext);
  if (context === undefined) {
    throw new Error("useBilling must be used within a BillingProvider");
  }
  return context;
}
