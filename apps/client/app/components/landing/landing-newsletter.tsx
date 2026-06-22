import { useState } from "react";
import { z } from "zod";
import {
  Loader2,
  Mail,
  AlertCircle,
  Send,
  MailCheck,
  CheckCircle2,
} from "lucide-react";
import { cn } from "@casbin/ui";
import { subscribeNewsletter } from "@/lib/email/newsletter-service";

const emailSchema = z
  .string()
  .trim()
  .nonempty({ message: "Email is required" })
  .email({ message: "Please enter a valid email" })
  .max(255, { message: "Email is too long" });

type Status = "idle" | "loading" | "success" | "error";

export function LandingNewsletter() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<Status>("idle");
  const [message, setMessage] = useState<string>("");
  const [submittedEmail, setSubmittedEmail] = useState<string>("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (status === "loading") return;

    const parsed = emailSchema.safeParse(email);
    if (!parsed.success) {
      setStatus("error");
      setMessage(parsed.error.errors[0]?.message ?? "Invalid email");
      return;
    }

    setStatus("loading");
    setMessage("");
    try {
      await subscribeNewsletter(parsed.data);
      setStatus("success");
      setSubmittedEmail(parsed.data);
      setEmail("");
    } catch (err) {
      setStatus("error");
      setMessage(err instanceof Error ? err.message : "Something went wrong");
    }
  };

  const isLoading = status === "loading";
  const isSuccess = status === "success";
  const isError = status === "error";

  // Success state — full panel takeover with "Check your inbox" message
  if (isSuccess) {
    return (
      <section className="border-border bg-surface/50 border-y">
        <div className="mx-auto max-w-7xl px-6 py-16 md:py-20">
          <div className="border-success/30 bg-card relative overflow-hidden rounded-2xl border p-8 md:p-12">
            <div className="bg-success/10 pointer-events-none absolute -top-16 -right-16 h-64 w-64 rounded-full blur-3xl" />
            <div className="bg-primary/10 pointer-events-none absolute -bottom-16 -left-16 h-64 w-64 rounded-full blur-3xl" />

            <div className="relative mx-auto max-w-xl text-center">
              <div className="bg-success/10 text-success mx-auto mb-5 inline-flex h-14 w-14 items-center justify-center rounded-full">
                <MailCheck className="h-7 w-7" />
              </div>
              <h3 className="text-foreground mb-3 text-2xl font-bold tracking-tight md:text-3xl">
                Check your inbox
              </h3>
              <p className="text-muted-foreground mb-2 text-sm md:text-base">
                We sent a confirmation link to{" "}
                <span className="text-foreground font-medium">
                  {submittedEmail}
                </span>
                .
              </p>
              <p className="text-muted-foreground text-sm">
                Click the link in the email to activate your subscription. The
                link expires in 24 hours.
              </p>

              <div className="mt-6 flex flex-col items-center justify-center gap-2 sm:flex-row">
                <button
                  onClick={() => {
                    setStatus("idle");
                    setSubmittedEmail("");
                    setMessage("");
                  }}
                  className="text-primary text-sm font-medium hover:underline"
                >
                  Use a different email
                </button>
              </div>

              <p className="text-muted-foreground mt-6 text-xs">
                Didn't get it? Check your spam folder, or wait a minute and try
                again.
              </p>
            </div>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section className="border-border bg-surface/50 border-y">
      <div className="mx-auto max-w-7xl px-6 py-16 md:py-20">
        <div className="border-border bg-card relative overflow-hidden rounded-2xl border p-8 md:p-12">
          <div className="bg-primary/10 pointer-events-none absolute -top-16 -right-16 h-64 w-64 rounded-full blur-3xl" />
          <div className="bg-accent/10 pointer-events-none absolute -bottom-16 -left-16 h-64 w-64 rounded-full blur-3xl" />

          <div className="relative grid grid-cols-1 items-center gap-8 md:grid-cols-2">
            <div>
              <div className="bg-primary/10 text-primary mb-4 inline-flex h-11 w-11 items-center justify-center rounded-lg">
                <Mail className="h-5 w-5" />
              </div>
              <h3 className="text-foreground mb-3 text-2xl font-bold tracking-tight md:text-3xl">
                Stay in the loop
              </h3>
              <p className="text-muted-foreground text-sm md:text-base">
                Get product updates, tips, and engineering deep-dives delivered
                straight to your inbox. No spam, unsubscribe anytime.
              </p>
            </div>

            <div>
              <form onSubmit={handleSubmit} noValidate className="space-y-3">
                <div className="flex flex-col gap-2 sm:flex-row">
                  <div className="relative flex-1">
                    <Mail className="text-muted-foreground pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2" />
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => {
                        setEmail(e.target.value);
                        if (status !== "idle") {
                          setStatus("idle");
                          setMessage("");
                        }
                      }}
                      placeholder="you@company.com"
                      disabled={isLoading || isSuccess}
                      maxLength={255}
                      aria-invalid={isError}
                      aria-describedby="newsletter-status"
                      className={cn(
                        "bg-background text-foreground placeholder:text-muted-foreground/70 h-11 w-full rounded-md border pr-3 pl-9 text-sm",
                        "focus:ring-ring transition-colors focus:border-transparent focus:ring-2 focus:outline-none",
                        "disabled:cursor-not-allowed disabled:opacity-60",
                        isError ? "border-danger" : "border-border",
                      )}
                    />
                  </div>
                  <button
                    type="submit"
                    disabled={isLoading || isSuccess}
                    className={cn(
                      "inline-flex h-11 items-center justify-center gap-2 rounded-md px-5 text-sm font-medium shadow-sm transition-all",
                      "bg-primary text-primary-foreground hover:bg-primary-hover",
                      "disabled:cursor-not-allowed disabled:opacity-70",
                    )}
                  >
                    {isLoading ? (
                      <>
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Subscribing...
                      </>
                    ) : isSuccess ? (
                      <>
                        <CheckCircle2 className="h-4 w-4" />
                        Subscribed
                      </>
                    ) : (
                      <>
                        Subscribe
                        <Send className="h-3.5 w-3.5" />
                      </>
                    )}
                  </button>
                </div>

                <div
                  id="newsletter-status"
                  role="status"
                  aria-live="polite"
                  className="min-h-[20px] text-xs"
                >
                  {isError && (
                    <span className="text-danger inline-flex items-center gap-1.5">
                      <AlertCircle className="h-3.5 w-3.5" />
                      {message}
                    </span>
                  )}
                  {!isError && (
                    <span className="text-muted-foreground">
                      We'll only email you about NexusOS — promise.
                    </span>
                  )}
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
