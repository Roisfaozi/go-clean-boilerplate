"use client";

import { Menubar as MenubarPrimitive } from "@base-ui/react/menubar";
import { Menu as MenuPrimitive } from "@base-ui/react/menu";
import { Check, ChevronRight, Circle } from "lucide-react";
import * as React from "react";

import { cn } from "./lib/utils";

function MenubarMenu({
	...props
}: React.ComponentProps<typeof MenuPrimitive.Root>) {
	return <MenuPrimitive.Root {...props} />;
}

function MenubarGroup({
	...props
}: React.ComponentProps<typeof MenuPrimitive.Group>) {
	return <MenuPrimitive.Group {...props} />;
}

function MenubarPortal({
	...props
}: React.ComponentProps<typeof MenuPrimitive.Portal>) {
	return <MenuPrimitive.Portal {...props} />;
}

function MenubarRadioGroup({
	...props
}: React.ComponentProps<typeof MenuPrimitive.RadioGroup>) {
	return <MenuPrimitive.RadioGroup {...props} />;
}

function MenubarSub({
	...props
}: React.ComponentProps<typeof MenuPrimitive.SubmenuRoot>) {
	return <MenuPrimitive.SubmenuRoot data-slot="menubar-sub" {...props} />;
}

const Menubar = React.forwardRef<
	React.ElementRef<typeof MenubarPrimitive>,
	React.ComponentProps<typeof MenubarPrimitive> & { loop?: boolean }
>(({ className, loop, loopFocus = loop, ...props }, ref) => (
	<MenubarPrimitive
		ref={ref}
		loopFocus={loopFocus}
		className={cn(
			"bg-background flex h-10 items-center space-x-1 rounded-md border p-1",
			className,
		)}
		{...props}
	/>
));
Menubar.displayName = "Menubar";

type WithAsChild = { asChild?: boolean };

const getRender = (asChild: boolean | undefined, children: React.ReactNode) =>
	asChild && React.isValidElement(children) ? children : undefined;

const MenubarTrigger = React.forwardRef<
	React.ElementRef<typeof MenuPrimitive.Trigger>,
	React.ComponentProps<typeof MenuPrimitive.Trigger> & WithAsChild
>(({ className, asChild, children, ...props }, ref) => (
	<MenuPrimitive.Trigger
		ref={ref as any}
		render={getRender(asChild, children)}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground data-popup-open:bg-accent data-popup-open:text-accent-foreground flex cursor-default items-center rounded-sm px-3 py-1.5 text-sm font-medium outline-hidden select-none",
			className,
		)}
		{...props}
	>
		{asChild ? undefined : children}
	</MenuPrimitive.Trigger>
));
MenubarTrigger.displayName = "MenubarTrigger";

const MenubarSubTrigger = React.forwardRef<
	React.ElementRef<typeof MenuPrimitive.SubmenuTrigger>,
	React.ComponentProps<typeof MenuPrimitive.SubmenuTrigger> & {
		inset?: boolean;
	} & WithAsChild
>(({ className, inset, asChild, children, ...props }, ref) => (
	<MenuPrimitive.SubmenuTrigger
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
	</MenuPrimitive.SubmenuTrigger>
));
MenubarSubTrigger.displayName = "MenubarSubTrigger";

type MenubarContentProps = Omit<
	React.ComponentProps<typeof MenuPrimitive.Popup>,
	"className"
> &
	Pick<
		React.ComponentProps<typeof MenuPrimitive.Positioner>,
		"align" | "alignOffset" | "side" | "sideOffset" | "collisionBoundary"
	> & {
		className?: string;
	};

const MenubarContent = React.forwardRef<
	React.ElementRef<typeof MenuPrimitive.Popup>,
	MenubarContentProps
>(
	(
		{
			className,
			align = "start",
			alignOffset = -4,
			sideOffset = 8,
			side,
			collisionBoundary,
			...props
		},
		ref,
	) => (
		<MenuPrimitive.Portal>
			<MenuPrimitive.Positioner
				align={align}
				alignOffset={alignOffset}
				side={side}
				sideOffset={sideOffset}
				collisionBoundary={collisionBoundary}
				className="isolate z-50 outline-none"
			>
				<MenuPrimitive.Popup
					ref={ref}
					className={cn(
						"bg-popover text-popover-foreground data-open:animate-in data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 min-w-48 origin-(--transform-origin) overflow-hidden rounded-md border p-1 shadow-md outline-none",
						className,
					)}
					{...props}
				/>
			</MenuPrimitive.Positioner>
		</MenuPrimitive.Portal>
	),
);
MenubarContent.displayName = "MenubarContent";

const MenubarSubContent = React.forwardRef<
	React.ElementRef<typeof MenuPrimitive.Popup>,
	MenubarContentProps
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
		<MenubarContent
			ref={ref}
			align={align}
			alignOffset={alignOffset}
			side={side}
			sideOffset={sideOffset}
			className={cn("min-w-32", className)}
			{...props}
		/>
	),
);
MenubarSubContent.displayName = "MenubarSubContent";

const MenubarItem = React.forwardRef<
	React.ElementRef<typeof MenuPrimitive.Item>,
	React.ComponentProps<typeof MenuPrimitive.Item> & {
		inset?: boolean;
	} & WithAsChild
>(({ className, inset, asChild, children, ...props }, ref) => (
	<MenuPrimitive.Item
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
	</MenuPrimitive.Item>
));
MenubarItem.displayName = "MenubarItem";

const MenubarCheckboxItem = React.forwardRef<
	React.ElementRef<typeof MenuPrimitive.CheckboxItem>,
	React.ComponentProps<typeof MenuPrimitive.CheckboxItem>
>(({ className, children, ...props }, ref) => (
	<MenuPrimitive.CheckboxItem
		ref={ref}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-hidden select-none data-disabled:pointer-events-none data-disabled:opacity-50",
			className,
		)}
		{...props}
	>
		<span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
			<MenuPrimitive.CheckboxItemIndicator>
				<Check className="h-4 w-4" />
			</MenuPrimitive.CheckboxItemIndicator>
		</span>
		{children}
	</MenuPrimitive.CheckboxItem>
));
MenubarCheckboxItem.displayName = "MenubarCheckboxItem";

const MenubarRadioItem = React.forwardRef<
	React.ElementRef<typeof MenuPrimitive.RadioItem>,
	React.ComponentProps<typeof MenuPrimitive.RadioItem>
>(({ className, children, ...props }, ref) => (
	<MenuPrimitive.RadioItem
		ref={ref}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground relative flex cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-hidden select-none data-disabled:pointer-events-none data-disabled:opacity-50",
			className,
		)}
		{...props}
	>
		<span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
			<MenuPrimitive.RadioItemIndicator>
				<Circle className="h-2 w-2 fill-current" />
			</MenuPrimitive.RadioItemIndicator>
		</span>
		{children}
	</MenuPrimitive.RadioItem>
));
MenubarRadioItem.displayName = "MenubarRadioItem";

const MenubarLabel = React.forwardRef<
	React.ElementRef<typeof MenuPrimitive.GroupLabel>,
	React.ComponentProps<typeof MenuPrimitive.GroupLabel> & { inset?: boolean }
>(({ className, inset, ...props }, ref) => (
	<MenuPrimitive.GroupLabel
		ref={ref}
		className={cn(
			"px-2 py-1.5 text-sm font-semibold",
			inset && "pl-8",
			className,
		)}
		{...props}
	/>
));
MenubarLabel.displayName = "MenubarLabel";

const MenubarSeparator = React.forwardRef<
	React.ElementRef<typeof MenuPrimitive.Separator>,
	React.ComponentProps<typeof MenuPrimitive.Separator>
>(({ className, ...props }, ref) => (
	<MenuPrimitive.Separator
		ref={ref}
		className={cn("bg-muted -mx-1 my-1 h-px", className)}
		{...props}
	/>
));
MenubarSeparator.displayName = "MenubarSeparator";

const MenubarShortcut = ({
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
MenubarShortcut.displayName = "MenubarShortcut";

export {
	Menubar,
	MenubarMenu,
	MenubarTrigger,
	MenubarContent,
	MenubarItem,
	MenubarSeparator,
	MenubarLabel,
	MenubarCheckboxItem,
	MenubarRadioGroup,
	MenubarRadioItem,
	MenubarPortal,
	MenubarSubContent,
	MenubarSubTrigger,
	MenubarGroup,
	MenubarSub,
	MenubarShortcut,
};
