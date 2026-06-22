import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "./lib/utils";
import { Loader2 } from "lucide-react";

const nexusButtonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap font-medium transition-colors duration-normal focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 rounded-md",
  {
    variants: {
      variant: {
        primary:
          "bg-primary text-primary-foreground hover:bg-primary-hover shadow-sm",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary-hover shadow-sm",
        outline:
          "border border-border bg-transparent hover:bg-surface-hover text-foreground",
        ghost: "hover:bg-surface-hover text-foreground",
        danger: "bg-danger text-danger-foreground hover:opacity-90 shadow-sm",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-btn px-[var(--button-padding-x)] text-body",
        sm: "h-8 px-3 text-small",
        lg: "h-12 px-8 text-body-lg",
        icon: "h-btn w-[var(--button-height)]",
      },
    },
    defaultVariants: {
      variant: "primary",
      size: "default",
    },
  },
);

export interface NexusButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof nexusButtonVariants> {
  asChild?: boolean;
  loading?: boolean;
}

const NexusButton = React.forwardRef<HTMLButtonElement, NexusButtonProps>(
  (
    {
      className,
      variant,
      size,
      asChild = false,
      loading,
      children,
      disabled,
      ...props
    },
    ref,
  ) => {
    const Comp = asChild ? Slot : "button";
    return (
      <Comp
        className={cn(nexusButtonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || loading}
        {...props}
      >
        {loading && <Loader2 className="h-4 w-4 animate-spin" />}
        {children}
      </Comp>
    );
  },
);
NexusButton.displayName = "NexusButton";

export { NexusButton, nexusButtonVariants };
