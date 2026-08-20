"use client";

import * as React from "react";
import { Checkbox as BaseCheckbox } from "@base-ui/react/checkbox";
import { Check } from "lucide-react";

import { cn } from "./lib/utils";

const Checkbox = React.forwardRef<
	React.ComponentRef<typeof BaseCheckbox.Root>,
	React.ComponentProps<typeof BaseCheckbox.Root>
>(({ className, ...props }, ref) => (
	<BaseCheckbox.Root
		ref={ref}
		className={cn(
			"peer border-primary ring-offset-background focus-visible:ring-ring data-checked:bg-primary data-checked:text-primary-foreground h-4 w-4 shrink-0 rounded-sm border focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none data-disabled:cursor-not-allowed data-disabled:opacity-50",
			className,
		)}
		{...props}
	>
		<BaseCheckbox.Indicator
			className={cn("flex items-center justify-center text-current")}
		>
			<Check className="h-4 w-4" />
		</BaseCheckbox.Indicator>
	</BaseCheckbox.Root>
));
Checkbox.displayName = "Checkbox";

export { Checkbox };
