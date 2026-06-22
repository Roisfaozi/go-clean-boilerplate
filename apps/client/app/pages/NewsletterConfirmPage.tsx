import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router";
import {
  CheckCircle2,
  AlertCircle,
  Loader2,
  Hexagon,
  ArrowRight,
} from "lucide-react";
import { confirmSubscription } from "@/lib/email/newsletter-service";

type Status = "loading" | "success" | "error";

export default function NewsletterConfirmPage() {
  const [params] = useSearchParams();
  const token = params.get("token") ?? "";

  const [status, setStatus] = useState<Status>("loading");
  const [message, setMessage] = useState<string>("");
  const [email, setEmail] = useState<string>("");

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await confirmSubscription(token);
        if (cancelled) return;
        setEmail(res.email ?? "");
        setStatus("success");
      } catch (err) {
        if (cancelled) return;
        setMessage(err instanceof Error ? err.message : "Confirmation failed");
        setStatus("error");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [token]);

  return (
    <div className="bg-background text-foreground flex min-h-screen flex-col">
      <header className="border-border border-b">
        <div className="mx-auto flex h-16 max-w-7xl items-center px-6">
          <Link to="/" className="flex items-center gap-2">
            <Hexagon className="text-primary h-7 w-7" />
            <span className="text-lg font-bold tracking-tight">NexusOS</span>
          </Link>
        </div>
      </header>

      <main className="flex flex-1 items-center justify-center px-6 py-16">
        <div className="border-border bg-card relative w-full max-w-md overflow-hidden rounded-2xl border p-8 text-center shadow-md md:p-10">
          <div className="bg-primary/10 pointer-events-none absolute -top-16 -right-16 h-56 w-56 rounded-full blur-3xl" />
          <div className="bg-accent/10 pointer-events-none absolute -bottom-16 -left-16 h-56 w-56 rounded-full blur-3xl" />

          <div className="relative">
            {status === "loading" && (
              <>
                <div className="bg-primary/10 text-primary mx-auto mb-5 inline-flex h-14 w-14 items-center justify-center rounded-full">
                  <Loader2 className="h-7 w-7 animate-spin" />
                </div>
                <h1 className="mb-2 text-xl font-bold tracking-tight md:text-2xl">
                  Confirming your subscription
                </h1>
                <p className="text-muted-foreground text-sm">
                  Hang tight, this only takes a moment...
                </p>
              </>
            )}

            {status === "success" && (
              <>
                <div className="bg-success/10 text-success mx-auto mb-5 inline-flex h-14 w-14 items-center justify-center rounded-full">
                  <CheckCircle2 className="h-7 w-7" />
                </div>
                <h1 className="mb-2 text-xl font-bold tracking-tight md:text-2xl">
                  You're subscribed!
                </h1>
                <p className="text-muted-foreground mb-6 text-sm">
                  {email ? (
                    <>
                      <span className="text-foreground font-medium">
                        {email}
                      </span>{" "}
                      is now confirmed. Welcome to the NexusOS newsletter.
                    </>
                  ) : (
                    <>
                      Your email is now confirmed. Welcome to the NexusOS
                      newsletter.
                    </>
                  )}
                </p>
                <Link
                  to="/"
                  className="bg-primary text-primary-foreground hover:bg-primary-hover inline-flex h-10 items-center justify-center gap-2 rounded-md px-5 text-sm font-medium shadow-sm transition-colors"
                >
                  Back to home
                  <ArrowRight className="h-4 w-4" />
                </Link>
              </>
            )}

            {status === "error" && (
              <>
                <div className="bg-danger/10 text-danger mx-auto mb-5 inline-flex h-14 w-14 items-center justify-center rounded-full">
                  <AlertCircle className="h-7 w-7" />
                </div>
                <h1 className="mb-2 text-xl font-bold tracking-tight md:text-2xl">
                  Confirmation failed
                </h1>
                <p className="text-muted-foreground mb-6 text-sm">{message}</p>
                <Link
                  to="/#newsletter"
                  className="border-border bg-background text-foreground hover:bg-muted inline-flex h-10 items-center justify-center gap-2 rounded-md border px-5 text-sm font-medium transition-colors"
                >
                  Try subscribing again
                </Link>
              </>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
