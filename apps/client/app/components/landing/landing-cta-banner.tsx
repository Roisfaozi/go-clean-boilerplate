import { Link } from "react-router";
import { ArrowRight } from "lucide-react";
import { cn } from "@casbin/ui";

interface LandingCtaBannerProps {
  variant?: "primary" | "surface";
  eyebrow?: string;
  title: string;
  description: string;
  primaryCta: { label: string; to: string };
  secondaryCta?: { label: string; to: string };
}

export function LandingCtaBanner({
  variant = "primary",
  eyebrow,
  title,
  description,
  primaryCta,
  secondaryCta,
}: LandingCtaBannerProps) {
  const isPrimary = variant === "primary";

  return (
    <section className="border-b border-border bg-background py-16 md:py-20">
      <div className="mx-auto max-w-7xl px-layout">
        <div
          className={cn(
            "relative overflow-hidden rounded-xl p-8 md:p-12 lg:p-14",
            isPrimary
              ? "bg-gradient-to-br from-primary to-accent text-primary-foreground shadow-xl"
              : "border border-border bg-surface text-foreground"
          )}
        >
          {isPrimary && (
            <>
              <div className="pointer-events-none absolute -right-16 -top-16 h-72 w-72 rounded-full bg-primary-foreground/10 blur-3xl" />
              <div className="pointer-events-none absolute -bottom-16 -left-16 h-72 w-72 rounded-full bg-primary-foreground/10 blur-3xl" />
            </>
          )}

          <div className="relative flex flex-col items-start justify-between gap-6 md:flex-row md:items-center">
            <div className="max-w-2xl">
              {eyebrow && (
                <span
                  className={cn(
                    "text-caption font-semibold uppercase tracking-wider",
                    isPrimary ? "text-primary-foreground/80" : "text-primary"
                  )}
                >
                  {eyebrow}
                </span>
              )}
              <h3
                className={cn(
                  "mt-2 text-2xl font-bold tracking-tight md:text-3xl",
                  isPrimary ? "text-primary-foreground" : "text-foreground"
                )}
              >
                {title}
              </h3>
              <p
                className={cn(
                  "mt-3 text-body-lg",
                  isPrimary ? "text-primary-foreground/85" : "text-muted-foreground"
                )}
              >
                {description}
              </p>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row md:shrink-0">
              <Link
                to={primaryCta.to}
                className={cn(
                  "group inline-flex h-btn items-center justify-center gap-2 rounded-md px-6 text-body font-medium shadow-sm transition-all",
                  isPrimary
                    ? "bg-background text-foreground hover:bg-surface"
                    : "bg-primary text-primary-foreground hover:bg-primary-hover"
                )}
              >
                {primaryCta.label}
                <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
              </Link>
              {secondaryCta && (
                <Link
                  to={secondaryCta.to}
                  className={cn(
                    "inline-flex h-btn items-center justify-center rounded-md border px-6 text-body font-medium transition-colors",
                    isPrimary
                      ? "border-primary-foreground/30 text-primary-foreground hover:bg-primary-foreground/10"
                      : "border-border text-foreground hover:bg-surface-hover"
                  )}
                >
                  {secondaryCta.label}
                </Link>
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}