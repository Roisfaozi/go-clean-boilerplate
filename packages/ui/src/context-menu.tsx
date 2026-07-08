"use client";

import { ContextMenu as ContextMenuPrimitive } from "@base-ui/react/context-menu";
import { Check, ChevronRight, Circle } from "lucide-react";
import * as React from "react";

import { cn } from "./lib/utils";

const ContextMenu = ContextMenuPrimitive.Root;

const ContextMenuGroup = ContextMenuPrimitive.Group;

const ContextMenuPortal = ContextMenuPrimitive.Portal;

const ContextMenuSub = ContextMenuPrimitive.SubmenuRoot;

const ContextMenuRadioGroup = ContextMenuPrimitive.RadioGroup;

type WithAsChild = {
	asChild?: boolean;
};

const getRender = (asChild: boolean | undefined, children: React.ReactNode) =>
	asChild && React.isValidElement(children) ? children : undefined;

const ContextMenuTrigger = React.forwardRef<
	React.ElementRef<typeof ContextMenuPrimitive.Trigger>,
	React.ComponentProps<typeof ContextMenuPrimitive.Trigger> & WithAsChild
>(({ asChild, children, ...props }, ref) => (
	<ContextMenuPrimitive.Trigger
		ref={ref}
		render={getRender(asChild, children)}
		{...props}
	>
		{asChild ? undefined : children}
	</ContextMenuPrimitive.Trigger>
));
ContextMenuTrigger.displayName = "ContextMenuTrigger";

const ContextMenuSubTrigger = React.forwardRef<
	React.ElementRef<typeof ContextMenuPrimitive.SubmenuTrigger>,
	React.ComponentProps<typeof ContextMenuPrimitive.SubmenuTrigger> & {
		inset?: boolean;
	} & WithAsChild
>(({ className, inset, asChild, children, ...props }, ref) => (
	<ContextMenuPrimitive.SubmenuTrigger
		ref={ref}
		render={getRender(asChild, children)}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground data-popup-open:bg-accent data-popup-open:text-accent-foreground flex cursor-default items-center rounded-sm px-2 py-1.5 text-sm outline-hidden select-none",
			inset && "pl-8",
			className,
		)}
		{...props}
	>
		{asChild ? undefined : (
			<>
				{children}
				<ChevronRight className="ml-auto h-4 w-4" />
			</>
		)}
	</ContextMenuPrimitive.SubmenuTrigger>
));
ContextMenuSubTrigger.displayName = "ContextMenuSubTrigger";

type ContextMenuContentProps = Omit<
	React.ComponentProps<typeof ContextMenuPrimitive.Popup>,
	"className"
> &
	Pick<
		React.ComponentProps<typeof ContextMenuPrimitive.Positioner>,
		"align" | "alignOffset" | "side" | "sideOffset" | "collisionBoundary"
	> & {
		className?: string;
	};

const ContextMenuContent = React.forwardRef<
	React.ElementRef<typeof ContextMenuPrimitive.Popup>,
	ContextMenuContentProps
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
		<ContextMenuPrimitive.Portal>
			<ContextMenuPrimitive.Positioner
				align={align}
				alignOffset={alignOffset}
				side={side}
				sideOffset={sideOffset}
				collisionBoundary={collisionBoundary}
				className="isolate z-50 outline-none"
			>
				<ContextMenuPrimitive.Popup
					ref={ref}
					className={cn(
						"bg-popover text-popover-foreground animate-in fade-in-80 data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 max-h-(--available-height) min-w-32 origin-(--transform-origin) overflow-x-hidden overflow-y-auto rounded-md border p-1 shadow-md outline-none",
						className,
					)}
					{...props}
				/>
			</ContextMenuPrimitive.Positioner>
		</ContextMenuPrimitive.Portal>
	),
);
ContextMenuContent.displayName = "ContextMenuContent";

const ContextMenuSubContent = React.forwardRef<
	React.ElementRef<typeof ContextMenuPrimitive.Popup>,
	ContextMenuContentProps
>(
	(
		{
			className,
			align = "start",
			alignOffset = 4,
			side = "right",
			sideOffset = 0,
			...props
		},
		ref,
	) => (
		<ContextMenuContent
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
ContextMenuSubContent.displayName = "ContextMenuSubContent";

const ContextMenuItem = React.forwardRef<
	React.ElementRef<typeof ContextMenuPrimitive.Item>,
	React.ComponentProps<typeof ContextMenuPrimitive.Item> & {
		inset?: boolean;
	} & WithAsChild
>(({ className, inset, asChild, children, ...props }, ref) => (
	<ContextMenuPrimitive.Item
		ref={ref}
		render={getRender(asChild, children)}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm px-2 py-1.5 text-sm outline-hidden select-none data-disabled:pointer-events-none data-disabled:opacity-50",
			inset && "pl-8",
			className,
		)}
		{...props}
	>
		{asChild ? undefined : children}
	</ContextMenuPrimitive.Item>
));
ContextMenuItem.displayName = "ContextMenuItem";

const ContextMenuCheckboxItem = React.forwardRef<
	React.ElementRef<typeof ContextMenuPrimitive.CheckboxItem>,
	React.ComponentProps<typeof ContextMenuPrimitive.CheckboxItem>
>(({ className, children, ...props }, ref) => (
	<ContextMenuPrimitive.CheckboxItem
		ref={ref}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-hidden select-none data-disabled:pointer-events-none data-disabled:opacity-50",
			className,
		)}
		{...props}
	>
		<span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
			<ContextMenuPrimitive.CheckboxItemIndicator>
				<Check className="h-4 w-4" />
			</ContextMenuPrimitive.CheckboxItemIndicator>
		</span>
		{children}
	</ContextMenuPrimitive.CheckboxItem>
));
ContextMenuCheckboxItem.displayName = "ContextMenuCheckboxItem";

const ContextMenuRadioItem = React.forwardRef<
	React.ElementRef<typeof ContextMenuPrimitive.RadioItem>,
	React.ComponentProps<typeof ContextMenuPrimitive.RadioItem>
>(({ className, children, ...props }, ref) => (
	<ContextMenuPrimitive.RadioItem
		ref={ref}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-hidden select-none data-disabled:pointer-events-none data-disabled:opacity-50",
			className,
		)}
		{...props}
	>
		<span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
			<ContextMenuPrimitive.RadioItemIndicator>
				<Circle className="h-2 w-2 fill-current" />
			</ContextMenuPrimitive.RadioItemIndicator>
		</span>
		{children}
	</ContextMenuPrimitive.RadioItem>
));
ContextMenuRadioItem.displayName = "ContextMenuRadioItem";

const ContextMenuLabel = React.forwardRef<
	React.ElementRef<typeof ContextMenuPrimitive.GroupLabel>,
	React.ComponentProps<typeof ContextMenuPrimitive.GroupLabel> & {
		inset?: boolean;
	}
>(({ className, inset, ...props }, ref) => (
	<ContextMenuPrimitive.GroupLabel
		ref={ref}
		className={cn(
			"text-foreground px-2 py-1.5 text-sm font-semibold",
			inset && "pl-8",
			className,
		)}
		{...props}
	/>
));
ContextMenuLabel.displayName = "ContextMenuLabel";

const ContextMenuSeparator = React.forwardRef<
	React.ElementRef<typeof ContextMenuPrimitive.Separator>,
	React.ComponentProps<typeof ContextMenuPrimitive.Separator>
>(({ className, ...props }, ref) => (
	<ContextMenuPrimitive.Separator
		ref={ref}
		className={cn("bg-border -mx-1 my-1 h-px", className)}
		{...props}
	/>
));
ContextMenuSeparator.displayName = "ContextMenuSeparator";

const ContextMenuShortcut = ({
	className,
	...props
}: React.HTMLAttributes<HTMLSpanElement>) => {
	return (
		<span
			className={cn(
				"text-muted-foreground ml-auto text-xs tracking-widest",
				className,
			)}
			{...props}
		/>
	);
};
ContextMenuShortcut.displayName = "ContextMenuShortcut";

export {
	ContextMenu,
	ContextMenuTrigger,
	ContextMenuContent,
	ContextMenuItem,
	ContextMenuCheckboxItem,
	ContextMenuRadioItem,
	ContextMenuLabel,
	ContextMenuSeparator,
	ContextMenuShortcut,
	ContextMenuGroup,
	ContextMenuPortal,
	ContextMenuSub,
	ContextMenuSubContent,
	ContextMenuSubTrigger,
	ContextMenuRadioGroup,
};
