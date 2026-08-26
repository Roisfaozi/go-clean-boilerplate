"use client";

import * as React from "react";
import { PreviewCard as PreviewCardPrimitive } from "@base-ui/react/preview-card";

import { cn } from "./lib/utils";

const HoverCard = PreviewCardPrimitive.Root;

const HoverCardTrigger = PreviewCardPrimitive.Trigger;

const HoverCardContent = React.forwardRef<
	React.ComponentRef<typeof PreviewCardPrimitive.Popup>,
	React.ComponentProps<typeof PreviewCardPrimitive.Popup> &
		Pick<
			React.ComponentProps<typeof PreviewCardPrimitive.Positioner>,
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
		<PreviewCardPrimitive.Portal>
			<PreviewCardPrimitive.Positioner
				align={align}
				sideOffset={sideOffset}
				alignOffset={alignOffset}
				side={side}
				collisionBoundary={collisionBoundary}
				className="z-50 outline-none"
			>
				<PreviewCardPrimitive.Popup
					ref={ref}
					className={cn(
						"bg-popover text-popover-foreground data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 origin-(--transform-origin) w-64 rounded-md border p-4 shadow-md outline-none",
						className,
					)}
					{...props}
				/>
			</PreviewCardPrimitive.Positioner>
		</PreviewCardPrimitive.Portal>
	),
);
HoverCardContent.displayName = "HoverCardContent";

export { HoverCard, HoverCardTrigger, HoverCardContent };
