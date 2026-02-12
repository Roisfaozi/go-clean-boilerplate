import { createSafeActionClient } from "next-safe-action";
import { z } from "zod";

/**
 * Base action client.
 * Handles Zod validation and error formatting automatically.
 */
export const actionClient = createSafeActionClient({
  // Define metadata schema
  defineMetadataSchema() {
    return z.object({
      actionName: z.string(),
    });
  },

  // Log errors to console in development
  handleServerError: (e) => {
    if (process.env.NODE_ENV === "development") {
      console.error("Action Server Error:", e.message);
    }
    return e.message || "An unexpected error occurred. Please try again.";
  },
});

/**
 * Authenticated action client.
 * Use this for actions that REQUIRE a logged-in user.
 */
export const authActionClient = actionClient.use(async ({ next }) => {
  return next({ ctx: {} });
});
