import { cn } from "@casbin/ui";
import { Check } from "lucide-react";

interface StepperFormProps {
  steps: { label: string; description?: string }[];
  currentStep: number;
  children: React.ReactNode;
  className?: string;
}

export function StepperForm({
  steps,
  currentStep,
  children,
  className,
}: StepperFormProps) {
  return (
    <div className={cn("space-y-8", className)}>
      {/* Step indicators */}
      <nav className="flex items-center justify-between">
        {steps.map((step, i) => {
          const status =
            i < currentStep
              ? "complete"
              : i === currentStep
                ? "current"
                : "upcoming";
          return (
            <div key={i} className="flex flex-1 items-center last:flex-none">
              <div className="flex items-center gap-3">
                <div
                  className={cn(
                    "text-caption flex h-8 w-8 shrink-0 items-center justify-center rounded-full font-semibold transition-colors",
                    status === "complete" &&
                      "bg-success text-success-foreground",
                    status === "current" &&
                      "bg-primary text-primary-foreground",
                    status === "upcoming" && "bg-muted text-muted-foreground",
                  )}
                >
                  {status === "complete" ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    i + 1
                  )}
                </div>
                <div className="hidden sm:block">
                  <p
                    className={cn(
                      "text-small font-medium",
                      status === "upcoming"
                        ? "text-muted-foreground"
                        : "text-foreground",
                    )}
                  >
                    {step.label}
                  </p>
                  {step.description && (
                    <p className="text-caption text-muted-foreground">
                      {step.description}
                    </p>
                  )}
                </div>
              </div>
              {i < steps.length - 1 && (
                <div
                  className={cn(
                    "mx-4 h-px flex-1",
                    i < currentStep ? "bg-success" : "bg-border",
                  )}
                />
              )}
            </div>
          );
        })}
      </nav>

      {/* Step content */}
      <div>{children}</div>
    </div>
  );
}
