"use client";

import * as React from "react";
import { Switch as BaseSwitch } from "@base-ui/react/switch";

import { cn } from "./lib/utils";

const Switch = React.forwardRef<
	React.ComponentRef<typeof BaseSwitch.Root>,
	React.ComponentProps<typeof BaseSwitch.Root>
>(({ className, ...props }, ref) => (
	<BaseSwitch.Root
		className={cn(
			"peer focus-visible:ring-ring focus-visible:ring-offset-background data-checked:bg-primary data-unchecked:bg-input inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none data-disabled:cursor-not-allowed data-disabled:opacity-50",
			className,
		)}
		{...props}
		ref={ref}
	>
		<BaseSwitch.Thumb
			className={cn(
				"bg-background pointer-events-none block h-5 w-5 rounded-full ring-0 shadow-lg transition-transform data-checked:translate-x-5 data-unchecked:translate-x-0",
			)}
		/>
	</BaseSwitch.Root>
));
Switch.displayName = "Switch";

export { Switch };
