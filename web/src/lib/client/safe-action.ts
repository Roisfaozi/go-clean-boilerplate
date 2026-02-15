import { createSafeActionClient } from "next-safe-action";
import { z } from "zod";
import { getCurrentSession } from "../server/auth";

export const actionClient = createSafeActionClient({
  defineMetadataSchema: () =>
    z.object({
      actionName: z.string(),
    }),
  handleServerError: (e) => {
    console.error("Action error:", e.message);
    return {
      success: false,
      message: e.message,
    };
  },
});

export const authActionClient = actionClient.use(async ({ next }) => {
  const { session, user } = await getCurrentSession();

  if (!session) {
    throw new Error("Session not found!");
  }

  return next({
    ctx: {
      user,
      sessionId: session.id,
    },
  });
});
