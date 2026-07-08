"use client";

import { Menu as DropdownMenuPrimitive } from "@base-ui/react/menu";
import { Check, ChevronRight, Circle } from "lucide-react";
import * as React from "react";

import { cn } from "./lib/utils";

const DropdownMenu = DropdownMenuPrimitive.Root;

const DropdownMenuGroup = DropdownMenuPrimitive.Group;

const DropdownMenuPortal = DropdownMenuPrimitive.Portal;

const DropdownMenuSub = DropdownMenuPrimitive.SubmenuRoot;

const DropdownMenuRadioGroup = DropdownMenuPrimitive.RadioGroup;

type WithAsChild = {
	asChild?: boolean;
};

const getRender = (asChild: boolean | undefined, children: React.ReactNode) =>
	asChild && React.isValidElement(children) ? children : undefined;

const DropdownMenuTrigger = React.forwardRef<
	React.ElementRef<typeof DropdownMenuPrimitive.Trigger>,
	React.ComponentProps<typeof DropdownMenuPrimitive.Trigger> & WithAsChild
>(({ asChild, children, ...props }, ref) => (
	<DropdownMenuPrimitive.Trigger
		ref={ref as any}
		render={getRender(asChild, children)}
		{...props}
	>
		{asChild ? undefined : children}
	</DropdownMenuPrimitive.Trigger>
));
DropdownMenuTrigger.displayName = "DropdownMenuTrigger";

const DropdownMenuSubTrigger = React.forwardRef<
	React.ElementRef<typeof DropdownMenuPrimitive.SubmenuTrigger>,
	React.ComponentProps<typeof DropdownMenuPrimitive.SubmenuTrigger> & {
		inset?: boolean;
	} & WithAsChild
>(({ className, inset, asChild, children, ...props }, ref) => (
	<DropdownMenuPrimitive.SubmenuTrigger
		ref={ref}
		render={getRender(asChild, children)}
		className={cn(
			"focus:bg-accent data-popup-open:bg-accent flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
			inset && "pl-8",
			className,
		)}
		{...props}
	>
		{asChild ? undefined : (
			<>
				{children}
				<ChevronRight className="ml-auto" />
			</>
		)}
	</DropdownMenuPrimitive.SubmenuTrigger>
));
DropdownMenuSubTrigger.displayName = "DropdownMenuSubTrigger";

type DropdownMenuContentProps = Omit<
	React.ComponentProps<typeof DropdownMenuPrimitive.Popup>,
	"className"
> &
	Pick<
		React.ComponentProps<typeof DropdownMenuPrimitive.Positioner>,
		| "align"
		| "alignOffset"
		| "side"
		| "sideOffset"
		| "collisionBoundary"
		| "collisionPadding"
	> & {
		className?: string;
	};

const DropdownMenuContent = React.forwardRef<
	React.ElementRef<typeof DropdownMenuPrimitive.Popup>,
	DropdownMenuContentProps
>(
	(
		{
			className,
			sideOffset = 4,
			align,
			alignOffset,
			side,
			collisionBoundary,
			collisionPadding,
			...props
		},
		ref,
	) => (
		<DropdownMenuPrimitive.Portal>
			<DropdownMenuPrimitive.Positioner
				sideOffset={sideOffset}
				align={align}
				alignOffset={alignOffset}
				side={side}
				collisionBoundary={collisionBoundary}
				collisionPadding={collisionPadding}
				className="isolate z-50 outline-none"
			>
				<DropdownMenuPrimitive.Popup
					ref={ref}
					className={cn(
						"bg-popover text-popover-foreground data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 max-h-(--available-height) min-w-32 origin-(--transform-origin) overflow-x-hidden overflow-y-auto rounded-md border p-1 shadow-md outline-none",
						className,
					)}
					{...props}
				/>
			</DropdownMenuPrimitive.Positioner>
		</DropdownMenuPrimitive.Portal>
	),
);
DropdownMenuContent.displayName = "DropdownMenuContent";

const DropdownMenuSubContent = React.forwardRef<
	React.ElementRef<typeof DropdownMenuPrimitive.Popup>,
	DropdownMenuContentProps
>(
	(
		{
			className,
			align = "start",
			alignOffset = -3,
			side = "right",
			sideOffset = 0,
			...props
		},
		ref,
	) => (
		<DropdownMenuContent
			ref={ref}
			align={align}
			alignOffset={alignOffset}
			side={side}
			sideOffset={sideOffset}
			className={cn("w-auto", className)}
			{...props}
		/>
	),
);
DropdownMenuSubContent.displayName = "DropdownMenuSubContent";

const DropdownMenuItem = React.forwardRef<
	React.ElementRef<typeof DropdownMenuPrimitive.Item>,
	React.ComponentProps<typeof DropdownMenuPrimitive.Item> & {
		inset?: boolean;
	} & WithAsChild
>(({ className, inset, asChild, children, ...props }, ref) => (
	<DropdownMenuPrimitive.Item
		ref={ref}
		render={getRender(asChild, children)}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden transition-colors select-none data-disabled:pointer-events-none data-disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
			inset && "pl-8",
			className,
		)}
		{...props}
	>
		{asChild ? undefined : children}
	</DropdownMenuPrimitive.Item>
));
DropdownMenuItem.displayName = "DropdownMenuItem";

const DropdownMenuCheckboxItem = React.forwardRef<
	React.ElementRef<typeof DropdownMenuPrimitive.CheckboxItem>,
	React.ComponentProps<typeof DropdownMenuPrimitive.CheckboxItem>
>(({ className, children, ...props }, ref) => (
	<DropdownMenuPrimitive.CheckboxItem
		ref={ref}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-hidden transition-colors select-none data-disabled:pointer-events-none data-disabled:opacity-50",
			className,
		)}
		{...props}
	>
		<span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
			<DropdownMenuPrimitive.CheckboxItemIndicator>
				<Check className="h-4 w-4" />
			</DropdownMenuPrimitive.CheckboxItemIndicator>
		</span>
		{children}
	</DropdownMenuPrimitive.CheckboxItem>
));
DropdownMenuCheckboxItem.displayName = "DropdownMenuCheckboxItem";

const DropdownMenuRadioItem = React.forwardRef<
	React.ElementRef<typeof DropdownMenuPrimitive.RadioItem>,
	React.ComponentProps<typeof DropdownMenuPrimitive.RadioItem>
>(({ className, children, ...props }, ref) => (
	<DropdownMenuPrimitive.RadioItem
		ref={ref}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-hidden transition-colors select-none data-disabled:pointer-events-none data-disabled:opacity-50",
			className,
		)}
		{...props}
	>
		<span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
			<DropdownMenuPrimitive.RadioItemIndicator>
				<Circle className="h-2 w-2 fill-current" />
			</DropdownMenuPrimitive.RadioItemIndicator>
		</span>
		{children}
	</DropdownMenuPrimitive.RadioItem>
));
DropdownMenuRadioItem.displayName = "DropdownMenuRadioItem";

const DropdownMenuLabel = React.forwardRef<
	React.ElementRef<typeof DropdownMenuPrimitive.GroupLabel>,
	React.ComponentProps<typeof DropdownMenuPrimitive.GroupLabel> & {
		inset?: boolean;
	}
>(({ className, inset, ...props }, ref) => (
	<DropdownMenuPrimitive.GroupLabel
		ref={ref}
		className={cn(
			"px-2 py-1.5 text-sm font-semibold",
			inset && "pl-8",
			className,
		)}
		{...props}
	/>
));
DropdownMenuLabel.displayName = "DropdownMenuLabel";

const DropdownMenuSeparator = React.forwardRef<
	React.ElementRef<typeof DropdownMenuPrimitive.Separator>,
	React.ComponentProps<typeof DropdownMenuPrimitive.Separator>
>(({ className, ...props }, ref) => (
	<DropdownMenuPrimitive.Separator
		ref={ref}
		className={cn("bg-muted -mx-1 my-1 h-px", className)}
		{...props}
	/>
));
DropdownMenuSeparator.displayName = "DropdownMenuSeparator";

const DropdownMenuShortcut = ({
	className,
	...props
}: React.HTMLAttributes<HTMLSpanElement>) => {
	return (
		<span
			className={cn("ml-auto text-xs tracking-widest opacity-60", className)}
			{...props}
		/>
	);
};
DropdownMenuShortcut.displayName = "DropdownMenuShortcut";

export {
	DropdownMenu,
	DropdownMenuTrigger,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuCheckboxItem,
	DropdownMenuRadioItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuShortcut,
	DropdownMenuGroup,
	DropdownMenuPortal,
	DropdownMenuSub,
	DropdownMenuSubContent,
	DropdownMenuSubTrigger,
	DropdownMenuRadioGroup,
};
