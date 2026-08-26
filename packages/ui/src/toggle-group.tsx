"use client";

import * as React from "react";
import { ToggleGroup as BaseToggleGroup } from "@base-ui/react/toggle-group";
import { Toggle as BaseToggle } from "@base-ui/react/toggle";
import type { VariantProps } from "class-variance-authority";

import { cn } from "./lib/utils";
import { toggleVariants } from "./toggle";

const ToggleGroupContext = React.createContext<
	VariantProps<typeof toggleVariants>
>({
	size: "default",
	variant: "default",
});

const ToggleGroup = React.forwardRef<
	React.ComponentRef<typeof BaseToggleGroup>,
	React.ComponentProps<typeof BaseToggleGroup> &
		VariantProps<typeof toggleVariants>
>(({ className, variant, size, children, ...props }, ref) => (
	<BaseToggleGroup
		ref={ref}
		className={cn("flex items-center justify-center gap-1", className)}
		{...props}
	>
		<ToggleGroupContext.Provider value={{ variant, size }}>
			{children}
		</ToggleGroupContext.Provider>
	</BaseToggleGroup>
));
ToggleGroup.displayName = "ToggleGroup";

const ToggleGroupItem = React.forwardRef<
	React.ComponentRef<typeof BaseToggle>,
	React.ComponentProps<typeof BaseToggle> &
		VariantProps<typeof toggleVariants> & { value: string }
>(({ className, children, variant, size, value, ...props }, ref) => {
	const context = React.useContext(ToggleGroupContext);

	return (
		<BaseToggle
			ref={ref}
			value={value}
			className={cn(
				toggleVariants({
					variant: context.variant || variant,
					size: context.size || size,
				}),
				className,
			)}
			{...props}
		>
			{children}
		</BaseToggle>
	);
});

ToggleGroupItem.displayName = "ToggleGroupItem";

export { ToggleGroup, ToggleGroupItem };
