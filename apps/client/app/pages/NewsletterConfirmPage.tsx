import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router";
import { CheckCircle2, AlertCircle, Loader2, Hexagon, ArrowRight } from "lucide-react";
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
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      <header className="border-b border-border">
        <div className="mx-auto flex h-16 max-w-7xl items-center px-6">
          <Link to="/" className="flex items-center gap-2">
            <Hexagon className="h-7 w-7 text-primary" />
            <span className="text-lg font-bold tracking-tight">NexusOS</span>
          </Link>
        </div>
      </header>

      <main className="flex flex-1 items-center justify-center px-6 py-16">
        <div className="relative w-full max-w-md overflow-hidden rounded-2xl border border-border bg-card p-8 text-center shadow-md md:p-10">
          <div className="pointer-events-none absolute -right-16 -top-16 h-56 w-56 rounded-full bg-primary/10 blur-3xl" />
          <div className="pointer-events-none absolute -bottom-16 -left-16 h-56 w-56 rounded-full bg-accent/10 blur-3xl" />

          <div className="relative">
            {status === "loading" && (
              <>
                <div className="mx-auto mb-5 inline-flex h-14 w-14 items-center justify-center rounded-full bg-primary/10 text-primary">
                  <Loader2 className="h-7 w-7 animate-spin" />
                </div>
                <h1 className="mb-2 text-xl font-bold tracking-tight md:text-2xl">
                  Confirming your subscription
                </h1>
                <p className="text-sm text-muted-foreground">Hang tight, this only takes a moment...</p>
              </>
            )}

            {status === "success" && (
              <>
                <div className="mx-auto mb-5 inline-flex h-14 w-14 items-center justify-center rounded-full bg-success/10 text-success">
                  <CheckCircle2 className="h-7 w-7" />
                </div>
                <h1 className="mb-2 text-xl font-bold tracking-tight md:text-2xl">
                  You're subscribed!
                </h1>
                <p className="mb-6 text-sm text-muted-foreground">
                  {email ? (
                    <>
                      <span className="font-medium text-foreground">{email}</span> is now
                      confirmed. Welcome to the NexusOS newsletter.
                    </>
                  ) : (
                    <>Your email is now confirmed. Welcome to the NexusOS newsletter.</>
                  )}
                </p>
                <Link
                  to="/"
                  className="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-primary px-5 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary-hover"
                >
                  Back to home
                  <ArrowRight className="h-4 w-4" />
                </Link>
              </>
            )}

            {status === "error" && (
              <>
                <div className="mx-auto mb-5 inline-flex h-14 w-14 items-center justify-center rounded-full bg-danger/10 text-danger">
                  <AlertCircle className="h-7 w-7" />
                </div>
                <h1 className="mb-2 text-xl font-bold tracking-tight md:text-2xl">
                  Confirmation failed
                </h1>
                <p className="mb-6 text-sm text-muted-foreground">{message}</p>
                <Link
                  to="/#newsletter"
                  className="inline-flex h-10 items-center justify-center gap-2 rounded-md border border-border bg-background px-5 text-sm font-medium text-foreground transition-colors hover:bg-muted"
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