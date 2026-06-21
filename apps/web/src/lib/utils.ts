import { generateRandomString, type RandomReader } from "@oslojs/crypto/random";
export { cn, nFormatter, hasFileNameSpaces, formatDate } from "@casbin/utils";

export class FreePlanLimitError extends Error {
  constructor(message?: string) {
    super(message || "Free plan limit reached. Please upgrade your plan.");
    this.name = "FreePlanLimitError";
  }
}

export function isRedirectError(error: unknown): boolean {
  return (
    error !== null &&
    typeof error === "object" &&
    "digest" in error &&
    typeof error.digest === "string" &&
    error.digest.includes("NEXT_REDIRECT")
  );
}

const alphanumeric =
  "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";

export function generateId(length = 10): string {
  const random: RandomReader = {
    read(bytes) {
      crypto.getRandomValues(bytes);
    },
  };
  return generateRandomString(random, alphanumeric, length);
}
