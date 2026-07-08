"use client";

import * as React from "react";
import { mergeProps } from "@base-ui/react/merge-props";
import { useRender } from "@base-ui/react/use-render";
import { cva, type VariantProps } from "class-variance-authority";
import { Loader2 } from "lucide-react";

import { cn } from "./lib/utils";

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
	extends useRender.ComponentProps<"button">,
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
			asChild,
			render,
			loading,
			children,
			disabled,
			...props
		},
		ref,
	) => {
		const renderElement =
			asChild && React.isValidElement(children) ? children : render;

		return useRender({
			ref,
			defaultTagName: "button",
			render: renderElement,
			props: mergeProps<"button">(
				{
					className: cn(nexusButtonVariants({ variant, size, className })),
					disabled: disabled || loading,
					children: renderElement ? undefined : (
						<>
							{loading && <Loader2 className="h-4 w-4 animate-spin" />}
							{children}
						</>
					),
				} as React.ComponentProps<"button">,
				props,
			),
		});
	},
);
NexusButton.displayName = "NexusButton";

export { NexusButton, nexusButtonVariants };
