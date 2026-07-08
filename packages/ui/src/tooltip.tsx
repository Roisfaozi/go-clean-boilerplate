"use client";

import * as React from "react";
import { Tooltip as TooltipPrimitive } from "@base-ui/react/tooltip";

import { cn } from "./lib/utils";

const TooltipProvider = TooltipPrimitive.Provider;

const Tooltip = TooltipPrimitive.Root;

const TooltipTrigger = TooltipPrimitive.Trigger;

const TooltipContent = React.forwardRef<
	React.ComponentRef<typeof TooltipPrimitive.Popup>,
	React.ComponentProps<typeof TooltipPrimitive.Popup> &
		Pick<
			React.ComponentProps<typeof TooltipPrimitive.Positioner>,
			"align" | "alignOffset" | "side" | "sideOffset" | "collisionBoundary"
		>
>(
	(
		{
			className,
			sideOffset = 4,
			align,
			alignOffset,
			side,
			collisionBoundary,
			...props
		},
		ref,
	) => (
		<TooltipPrimitive.Portal>
			<TooltipPrimitive.Positioner
				sideOffset={sideOffset}
				align={align}
				alignOffset={alignOffset}
				side={side}
				collisionBoundary={collisionBoundary}
				className="z-50 outline-none"
			>
				<TooltipPrimitive.Popup
					ref={ref}
					className={cn(
						"bg-popover text-popover-foreground animate-in fade-in-0 zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 origin-[var(--transform-origin)] overflow-hidden rounded-md border px-3 py-1.5 text-sm shadow-md outline-none",
						className,
					)}
					{...props}
				/>
			</TooltipPrimitive.Positioner>
		</TooltipPrimitive.Portal>
	),
);
TooltipContent.displayName = "TooltipContent";

export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider };
