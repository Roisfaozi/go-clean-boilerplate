"use client";

import * as React from "react";
import { Slider as BaseSlider } from "@base-ui/react/slider";

import { cn } from "./lib/utils";

const Slider = React.forwardRef<
	React.ComponentRef<typeof BaseSlider.Root>,
	React.ComponentProps<typeof BaseSlider.Root>
>(({ className, ...props }, ref) => (
	<BaseSlider.Root
		ref={ref}
		className={cn(
			"relative flex w-full touch-none items-center select-none",
			className,
		)}
		{...props}
	>
		<BaseSlider.Control className="flex w-full items-center">
			<BaseSlider.Track className="bg-secondary relative h-2 w-full grow overflow-hidden rounded-full">
				<BaseSlider.Indicator className="bg-primary absolute h-full" />
			</BaseSlider.Track>
			<BaseSlider.Thumb className="border-primary bg-background ring-offset-background focus-visible:ring-ring block h-5 w-5 rounded-full border-2 transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none data-disabled:pointer-events-none data-disabled:opacity-50" />
		</BaseSlider.Control>
	</BaseSlider.Root>
));
Slider.displayName = "Slider";

export { Slider };
