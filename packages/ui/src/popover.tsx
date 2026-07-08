"use client";

import * as React from "react";
import { Popover as PopoverPrimitive } from "@base-ui/react/popover";

import { cn } from "./lib/utils";

const Popover = PopoverPrimitive.Root;

const PopoverTrigger = PopoverPrimitive.Trigger;

const PopoverContent = React.forwardRef<
	React.ComponentRef<typeof PopoverPrimitive.Popup>,
	React.ComponentProps<typeof PopoverPrimitive.Popup> &
		Pick<
			React.ComponentProps<typeof PopoverPrimitive.Positioner>,
			"align" | "alignOffset" | "side" | "sideOffset" | "collisionBoundary"
		>
>(
	(
		{
			className,
			align = "center",
			sideOffset = 4,
			alignOffset,
			side,
			collisionBoundary,
			...props
		},
		ref,
	) => (
		<PopoverPrimitive.Portal>
			<PopoverPrimitive.Positioner
				align={align}
				sideOffset={sideOffset}
				alignOffset={alignOffset}
				side={side}
				collisionBoundary={collisionBoundary}
				className="z-50 outline-none"
			>
				<PopoverPrimitive.Popup
					ref={ref}
					className={cn(
						"bg-popover text-popover-foreground data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 origin-[var(--transform-origin)] w-72 rounded-md border p-4 shadow-md outline-none",
						className,
					)}
					{...props}
				/>
			</PopoverPrimitive.Positioner>
		</PopoverPrimitive.Portal>
	),
);
PopoverContent.displayName = "PopoverContent";

export { Popover, PopoverTrigger, PopoverContent };
