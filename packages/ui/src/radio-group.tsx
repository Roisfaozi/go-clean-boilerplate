"use client";

import * as React from "react";
import { RadioGroup as BaseRadioGroup } from "@base-ui/react/radio-group";
import { Radio as BaseRadio } from "@base-ui/react/radio";
import { Circle } from "lucide-react";

import { cn } from "./lib/utils";

const RadioGroup = React.forwardRef<
	React.ComponentRef<typeof BaseRadioGroup>,
	React.ComponentProps<typeof BaseRadioGroup>
>(({ className, ...props }, ref) => {
	return (
		<BaseRadioGroup
			className={cn("grid gap-2", className)}
			{...props}
			ref={ref}
		/>
	);
});
RadioGroup.displayName = "RadioGroup";

const RadioGroupItem = React.forwardRef<
	React.ComponentRef<typeof BaseRadio.Root>,
	React.ComponentProps<typeof BaseRadio.Root>
>(({ className, ...props }, ref) => {
	return (
		<BaseRadio.Root
			ref={ref}
			className={cn(
				"border-primary text-primary ring-offset-background focus-visible:ring-ring aspect-square h-4 w-4 rounded-full border focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 data-disabled:cursor-not-allowed data-disabled:opacity-50",
				className,
			)}
			{...props}
		>
			<BaseRadio.Indicator className="flex items-center justify-center">
				<Circle className="h-2.5 w-2.5 fill-current text-current" />
			</BaseRadio.Indicator>
		</BaseRadio.Root>
	);
});
RadioGroupItem.displayName = "RadioGroupItem";

export { RadioGroup, RadioGroupItem };
