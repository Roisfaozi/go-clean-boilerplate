/**
 * Newsletter subscription service.
 *
 * Adapter pattern — swap providers without touching UI code.
 * Configure via env: VITE_NEWSLETTER_PROVIDER = "mock" | "resend" | "custom"
 *                    VITE_NEWSLETTER_ENDPOINT = full URL for "resend" / "custom"
 *
 * The expected endpoint contract (POST application/json):
 *   Request:  { email: string, confirmUrl: string }
 *   Response: { ok: true, token?: string } | { ok: false, error: string }
 *
 * The endpoint is responsible for:
 *   1. Generating a confirmation token
 *   2. Persisting { email, token, confirmed: false } in your store
 *   3. Sending the welcome/confirmation email containing `${confirmUrl}?token=${token}`
 *
 * Easy migration paths:
 *   - Resend: build a tiny serverless route (Vercel / Cloudflare / Express) that
 *     calls Resend's POST /emails with your template, then point
 *     VITE_NEWSLETTER_ENDPOINT at it.
 *   - In-house: same shape, point at your own backend.
 */

export interface SubscribeResult {
  ok: true;
  /** Only present in mock mode for local testing */
  devToken?: string;
}

export interface ConfirmResult {
  ok: true;
  email?: string;
}

type Provider = "mock" | "resend" | "custom";

const PROVIDER: Provider =
  (import.meta.env.VITE_NEWSLETTER_PROVIDER as Provider) || "mock";
const ENDPOINT: string | undefined = import.meta.env.VITE_NEWSLETTER_ENDPOINT;

const MOCK_STORE_KEY = "newsletter:mock-subscribers";

interface MockRecord {
  email: string;
  token: string;
  confirmed: boolean;
  createdAt: string;
}

function readMockStore(): MockRecord[] {
  try {
    const raw = localStorage.getItem(MOCK_STORE_KEY);
    return raw ? (JSON.parse(raw) as MockRecord[]) : [];
  } catch {
    return [];
  }
}

function writeMockStore(records: MockRecord[]) {
  localStorage.setItem(MOCK_STORE_KEY, JSON.stringify(records));
}

function buildConfirmUrl(): string {
  return `${window.location.origin}/newsletter/confirm`;
}

function generateToken(): string {
  // crypto.randomUUID is available in modern browsers
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID().replace(/-/g, "");
  }
  return Math.random().toString(36).slice(2) + Date.now().toString(36);
}

/**
 * Subscribe an email. Returns once the welcome email has been queued.
 * Throws an Error with a user-safe message on failure.
 */
export async function subscribeNewsletter(
  email: string,
): Promise<SubscribeResult> {
  const normalized = email.trim().toLowerCase();

  // Test hook: any address containing "fail" forces an error path
  if (normalized.includes("fail")) {
    await new Promise((r) => setTimeout(r, 800));
    throw new Error("This email could not be subscribed. Try another.");
  }

  if (PROVIDER === "mock") {
    await new Promise((r) => setTimeout(r, 1000));
    const store = readMockStore();
    const existing = store.find((r) => r.email === normalized);
    if (existing && existing.confirmed) {
      throw new Error("This email is already subscribed.");
    }
    const token = existing?.token ?? generateToken();
    if (!existing) {
      store.push({
        email: normalized,
        token,
        confirmed: false,
        createdAt: new Date().toISOString(),
      });
      writeMockStore(store);
    }
    // Useful for local dev: surface the would-be confirm link
    // eslint-disable-next-line no-console
    console.info(
      `[newsletter:mock] confirm link → ${buildConfirmUrl()}?token=${token}`,
    );
    return { ok: true, devToken: token };
  }

  // Real provider — POST to the configured endpoint
  if (!ENDPOINT) {
    throw new Error(
      "Newsletter endpoint is not configured. Set VITE_NEWSLETTER_ENDPOINT.",
    );
  }

  const res = await fetch(ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: normalized, confirmUrl: buildConfirmUrl() }),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok || data?.ok === false) {
    throw new Error(data?.error ?? "Could not subscribe right now. Try again.");
  }
  return { ok: true };
}

/**
 * Confirm a subscription token. Used by the /newsletter/confirm page.
 */
export async function confirmSubscription(
  token: string,
): Promise<ConfirmResult> {
  const trimmed = token.trim();
  if (!trimmed) throw new Error("Missing confirmation token.");

  if (PROVIDER === "mock") {
    await new Promise((r) => setTimeout(r, 700));
    const store = readMockStore();
    const record = store.find((r) => r.token === trimmed);
    if (!record) {
      throw new Error("This confirmation link is invalid or has expired.");
    }
    record.confirmed = true;
    writeMockStore(store);
    return { ok: true, email: record.email };
  }

  if (!ENDPOINT) {
    throw new Error("Newsletter endpoint is not configured.");
  }

  const url = `${ENDPOINT.replace(/\/+$/, "")}/confirm`;
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ token: trimmed }),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok || data?.ok === false) {
    throw new Error(data?.error ?? "Could not confirm your subscription.");
  }
  return { ok: true, email: data?.email };
}
