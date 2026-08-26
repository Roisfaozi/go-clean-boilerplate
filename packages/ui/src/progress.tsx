"use client";

import * as React from "react";
import { Progress } from "@base-ui/react/progress";

import { cn } from "./lib/utils";

const ProgressRoot = React.forwardRef<
	React.ComponentRef<typeof Progress.Root>,
	React.ComponentProps<typeof Progress.Root>
>(({ className, ...props }, ref) => (
	<Progress.Root
		ref={ref}
		className={cn(
			"bg-secondary relative h-4 w-full overflow-hidden rounded-full",
			className,
		)}
		{...props}
	>
		<Progress.Track className="h-full w-full">
			<Progress.Indicator className="bg-primary h-full w-full transition-all">
				<Progress.Value />
			</Progress.Indicator>
		</Progress.Track>
	</Progress.Root>
));
ProgressRoot.displayName = "Progress";

export { ProgressRoot as Progress };
