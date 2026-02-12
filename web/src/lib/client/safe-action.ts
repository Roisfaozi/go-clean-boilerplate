"use client";

import { createSafeActionClient } from "next-safe-action";

/**
 * Base action client.
 * Handles Zod validation and error formatting automatically.
 */
export const actionClient = createSafeActionClient({
  // Log errors to console in development
  handleServerErrorLog: (e) => {
    if (process.env.NODE_ENV === "development") {
      console.error("Action Server Error:", e.message);
    }
  },
  
  // Custom error message for production
  handleReturnedServerError: (e) => {
    return e.message || "An unexpected error occurred. Please try again.";
  }
});

/**
 * Authenticated action client.
 * Use this for actions that REQUIRE a logged-in user.
 * (Logic placeholder - usually checks for session/cookie)
 */
export const authActionClient = actionClient.use(async ({ next }) => {
  // In a real implementation, we could verify the session here on the server
  // For now, it inherits the base client behavior.
  return next({ ctx: {} });
});
