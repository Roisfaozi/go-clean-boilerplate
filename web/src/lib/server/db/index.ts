// Placeholder for prisma db
export const prisma = {
  user: {
    findUnique: async (args?: any) => ({ id: "placeholder-user-id" }) as any,
    create: async (args?: any) => ({ id: "placeholder-user-id" }) as any,
    update: async (args?: any) => ({ id: "placeholder-user-id" }) as any,
    findFirst: async (args?: any) => ({ id: "placeholder-user-id" }) as any,
    upsert: async (args?: any) => ({ id: "placeholder-user-id" }) as any,
  },
  session: {
    create: async (args?: any) => ({ id: "placeholder-session-id" }) as any,
    delete: async (args?: any) => null,
  },
  project: {
    create: async (args?: any) => ({ id: "placeholder-project-id" }) as any,
    count: async (args?: any) => 0,
    findMany: async (args?: any) => [],
    findFirst: async (args?: any) => ({ id: "placeholder-project-id" }) as any,
    update: async (args?: any) => ({ id: "placeholder-project-id" }) as any,
    delete: async (args?: any) => ({ id: "placeholder-project-id" }) as any,
  },
};
