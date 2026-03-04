// Placeholder for payment service
export const stripe = {
  subscriptions: {
    retrieve: async (id: string) =>
      ({
        id: "placeholder-sub-id",
        customer: "placeholder-cust-id",
        cancel_at_period_end: false,
        items: {
          data: [{ price: { id: "placeholder-price-id" } }],
        },
      }) as any,
  },
  billingPortal: {
    sessions: {
      create: async (args: any) => ({ url: "" }),
    },
  },
  checkout: {
    sessions: {
      create: async (args: any) => ({ url: "" }),
    },
  },
  webhooks: {
    constructEvent: (body: any, sig: any, secret: string) => ({}) as any,
  },
};

export const getUserSubscriptionPlan = async (userId: string) => ({
  isPro: false,
  stripeSubscriptionId: null,
  stripeCurrentPeriodEnd: null,
  stripeCustomerId: null,
});

export const proPlan = { name: "Pro", description: "Pro Plan", price: 10 };
