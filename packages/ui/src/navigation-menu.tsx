import { NavigationMenu as NavigationMenuPrimitive } from "@base-ui/react/navigation-menu";
import { cva } from "class-variance-authority";
import { ChevronDown } from "lucide-react";
import * as React from "react";

import { cn } from "./lib/utils";

const NavigationMenu = React.forwardRef<
	React.ElementRef<typeof NavigationMenuPrimitive.Root>,
	React.ComponentProps<typeof NavigationMenuPrimitive.Root>
>(({ className, children, ...props }, ref) => (
	<NavigationMenuPrimitive.Root
		ref={ref}
		className={cn(
			"relative z-10 flex max-w-max flex-1 items-center justify-center",
			className,
		)}
		{...props}
	>
		{children}
		<NavigationMenuViewport />
	</NavigationMenuPrimitive.Root>
));
NavigationMenu.displayName = "NavigationMenu";

const NavigationMenuList = React.forwardRef<
	React.ElementRef<typeof NavigationMenuPrimitive.List>,
	React.ComponentProps<typeof NavigationMenuPrimitive.List>
>(({ className, ...props }, ref) => (
	<NavigationMenuPrimitive.List
		ref={ref}
		className={cn(
			"group flex flex-1 list-none items-center justify-center space-x-1",
			className,
		)}
		{...props}
	/>
));
NavigationMenuList.displayName = "NavigationMenuList";

const NavigationMenuItem = NavigationMenuPrimitive.Item;

const navigationMenuTriggerStyle = cva(
	"group inline-flex h-10 w-max items-center justify-center rounded-md bg-background px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-hidden disabled:pointer-events-none disabled:opacity-50 data-popup-open:text-accent-foreground data-popup-open:bg-accent/50 data-popup-open:hover:bg-accent data-popup-open:focus:bg-accent",
);

type WithAsChild = { asChild?: boolean };
const getRender = (asChild: boolean | undefined, children: React.ReactNode) =>
	asChild && React.isValidElement(children) ? children : undefined;

const NavigationMenuTrigger = React.forwardRef<
	React.ElementRef<typeof NavigationMenuPrimitive.Trigger>,
	React.ComponentProps<typeof NavigationMenuPrimitive.Trigger> & WithAsChild
>(({ className, asChild, children, ...props }, ref) => (
	<NavigationMenuPrimitive.Trigger
		ref={ref}
		render={getRender(asChild, children)}
		className={cn(navigationMenuTriggerStyle(), "group", className)}
		{...props}
	>
		{asChild ? undefined : (
			<>
				{children}{" "}
				<NavigationMenuPrimitive.Icon
					render={
						<ChevronDown
							className="relative top-px ml-1 h-3 w-3 transition duration-200 group-data-popup-open:rotate-180"
							aria-hidden="true"
						/>
					}
				/>
			</>
		)}
	</NavigationMenuPrimitive.Trigger>
));
NavigationMenuTrigger.displayName = "NavigationMenuTrigger";

const NavigationMenuContent = React.forwardRef<
	React.ElementRef<typeof NavigationMenuPrimitive.Content>,
	React.ComponentProps<typeof NavigationMenuPrimitive.Content>
>(({ className, ...props }, ref) => (
	<NavigationMenuPrimitive.Content
		ref={ref}
		className={cn(
			"data-[activation-direction=right]:animate-in data-[activation-direction=left]:animate-in data-[activation-direction=right]:slide-in-from-right-52 data-[activation-direction=left]:slide-in-from-left-52 top-0 left-0 w-full md:absolute md:w-auto",
			className,
		)}
		{...props}
	/>
));
NavigationMenuContent.displayName = "NavigationMenuContent";

const NavigationMenuLink = NavigationMenuPrimitive.Link;

const NavigationMenuViewport = React.forwardRef<
	React.ElementRef<typeof NavigationMenuPrimitive.Viewport>,
	React.ComponentProps<typeof NavigationMenuPrimitive.Viewport> &
		Pick<
			React.ComponentProps<typeof NavigationMenuPrimitive.Positioner>,
			"align" | "alignOffset" | "side" | "sideOffset" | "collisionBoundary"
		>
>(
	(
		{
			className,
			align,
			alignOffset,
			side,
			sideOffset,
			collisionBoundary,
			...props
		},
		ref,
	) => (
		<NavigationMenuPrimitive.Portal>
			<NavigationMenuPrimitive.Positioner
				align={align}
				alignOffset={alignOffset}
				side={side}
				sideOffset={sideOffset}
				collisionBoundary={collisionBoundary}
				className="absolute top-full left-0 isolate z-50 flex justify-center outline-none"
			>
				<NavigationMenuPrimitive.Popup className="origin-top-center bg-popover text-popover-foreground data-open:animate-in data-closed:animate-out data-closed:zoom-out-95 data-open:zoom-in-90 relative mt-1.5 overflow-hidden rounded-md border shadow-lg">
					<NavigationMenuPrimitive.Viewport
						className={cn(
							"h-(--popup-height) w-full md:w-(--popup-width)",
							className,
						)}
						ref={ref}
						{...props}
					/>
				</NavigationMenuPrimitive.Popup>
			</NavigationMenuPrimitive.Positioner>
		</NavigationMenuPrimitive.Portal>
	),
);
NavigationMenuViewport.displayName = "NavigationMenuViewport";

const NavigationMenuIndicator = React.forwardRef<
	React.ElementRef<typeof NavigationMenuPrimitive.Icon>,
	React.ComponentProps<typeof NavigationMenuPrimitive.Icon>
>(({ className, ...props }, ref) => (
	<NavigationMenuPrimitive.Icon
		ref={ref}
		className={cn(
			"data-popup-open:animate-in top-full z-1 flex h-1.5 items-end justify-center overflow-hidden",
			className,
		)}
		{...props}
	>
		<div className="bg-border relative top-[60%] h-2 w-2 rotate-45 rounded-tl-sm shadow-md" />
	</NavigationMenuPrimitive.Icon>
));
NavigationMenuIndicator.displayName = "NavigationMenuIndicator";

export {
	navigationMenuTriggerStyle,
	NavigationMenu,
	NavigationMenuList,
	NavigationMenuItem,
	NavigationMenuContent,
	NavigationMenuTrigger,
	NavigationMenuLink,
	NavigationMenuIndicator,
	NavigationMenuViewport,
};
