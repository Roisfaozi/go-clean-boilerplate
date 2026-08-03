"use client";

import { Select as SelectPrimitive } from "@base-ui/react/select";
import { Check, ChevronDown, ChevronUp } from "lucide-react";
import * as React from "react";

import { cn } from "./lib/utils";

const Select = SelectPrimitive.Root;

const SelectGroup = SelectPrimitive.Group;

const SelectValue = SelectPrimitive.Value;

type WithAsChild = { asChild?: boolean };
const getRender = (asChild: boolean | undefined, children: React.ReactNode) =>
	asChild && React.isValidElement(children) ? children : undefined;

const SelectTrigger = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.Trigger>,
	React.ComponentProps<typeof SelectPrimitive.Trigger> & WithAsChild
>(({ className, asChild, children, ...props }, ref) => (
	<SelectPrimitive.Trigger
		ref={ref}
		render={getRender(asChild, children)}
		className={cn(
			"border-input bg-background ring-offset-background focus:ring-ring data-placeholder:text-muted-foreground flex w-full items-center justify-between border px-3 py-2 text-sm focus:ring-2 focus:ring-offset-2 focus:outline-hidden disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1",
			"h-input rounded-md text-body",
			className,
		)}
		{...props}
	>
		{asChild ? undefined : (
			<>
				{children}
				<SelectPrimitive.Icon
					render={<ChevronDown className="h-4 w-4 opacity-50" />}
				/>
			</>
		)}
	</SelectPrimitive.Trigger>
));
SelectTrigger.displayName = "SelectTrigger";

const SelectScrollUpButton = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.ScrollUpArrow>,
	React.ComponentProps<typeof SelectPrimitive.ScrollUpArrow>
>(({ className, ...props }, ref) => (
	<SelectPrimitive.ScrollUpArrow
		ref={ref}
		className={cn(
			"flex cursor-default items-center justify-center py-1 top-0 w-full",
			className,
		)}
		{...props}
	>
		<ChevronUp className="h-4 w-4" />
	</SelectPrimitive.ScrollUpArrow>
));
SelectScrollUpButton.displayName = "SelectScrollUpButton";

const SelectScrollDownButton = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.ScrollDownArrow>,
	React.ComponentProps<typeof SelectPrimitive.ScrollDownArrow>
>(({ className, ...props }, ref) => (
	<SelectPrimitive.ScrollDownArrow
		ref={ref}
		className={cn(
			"flex cursor-default items-center justify-center py-1 bottom-0 w-full",
			className,
		)}
		{...props}
	>
		<ChevronDown className="h-4 w-4" />
	</SelectPrimitive.ScrollDownArrow>
));
SelectScrollDownButton.displayName = "SelectScrollDownButton";

type SelectContentProps = Omit<
	React.ComponentProps<typeof SelectPrimitive.Popup>,
	"className"
> &
	Pick<
		React.ComponentProps<typeof SelectPrimitive.Positioner>,
		| "align"
		| "alignOffset"
		| "side"
		| "sideOffset"
		| "alignItemWithTrigger"
		| "collisionBoundary"
	> & {
		className?: string;
	};

const SelectContent = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.Popup>,
	SelectContentProps
>(
	(
		{
			className,
			children,
			alignItemWithTrigger = true,
			align,
			alignOffset,
			side,
			sideOffset = 4,
			collisionBoundary,
			...props
		},
		ref,
	) => (
		<SelectPrimitive.Portal>
			<SelectPrimitive.Positioner
				alignItemWithTrigger={alignItemWithTrigger}
				align={align}
				alignOffset={alignOffset}
				side={side}
				sideOffset={sideOffset}
				collisionBoundary={collisionBoundary}
			>
				<SelectPrimitive.Popup
					ref={ref}
					className={cn(
						"bg-popover text-popover-foreground data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 relative isolate z-50 max-h-(--available-height) min-w-32 origin-(--transform-origin) overflow-x-hidden overflow-y-auto rounded-md border shadow-md",
						!alignItemWithTrigger &&
							"data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1",
						className,
					)}
					{...props}
				>
					<SelectScrollUpButton />
					<SelectPrimitive.List
						className={cn(
							"p-1",
							!alignItemWithTrigger &&
								"h-(--anchor-height) w-full min-w-(--anchor-width)",
						)}
					>
						{children}
					</SelectPrimitive.List>
					<SelectScrollDownButton />
				</SelectPrimitive.Popup>
			</SelectPrimitive.Positioner>
		</SelectPrimitive.Portal>
	),
);
SelectContent.displayName = "SelectContent";

const SelectLabel = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.GroupLabel>,
	React.ComponentProps<typeof SelectPrimitive.GroupLabel>
>(({ className, ...props }, ref) => (
	<SelectPrimitive.GroupLabel
		ref={ref}
		className={cn("py-1.5 pr-2 pl-8 text-sm font-semibold", className)}
		{...props}
	/>
));
SelectLabel.displayName = "SelectLabel";

const SelectItem = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.Item>,
	React.ComponentProps<typeof SelectPrimitive.Item> & WithAsChild
>(({ className, asChild, children, ...props }, ref) => (
	<SelectPrimitive.Item
		ref={ref}
		render={getRender(asChild, children)}
		className={cn(
			"focus:bg-accent focus:text-accent-foreground relative flex w-full cursor-default items-center rounded-sm py-1.5 pr-2 pl-8 text-sm outline-hidden select-none data-disabled:pointer-events-none data-disabled:opacity-50",
			className,
		)}
		{...props}
	>
		{asChild ? undefined : (
			<>
				<span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
					<SelectPrimitive.ItemIndicator>
						<Check className="h-4 w-4" />
					</SelectPrimitive.ItemIndicator>
				</span>
				<SelectPrimitive.ItemText className="shrink-0 whitespace-nowrap">
					{children}
				</SelectPrimitive.ItemText>
			</>
		)}
	</SelectPrimitive.Item>
));
SelectItem.displayName = "SelectItem";

const SelectSeparator = React.forwardRef<
	React.ElementRef<typeof SelectPrimitive.Separator>,
	React.ComponentProps<typeof SelectPrimitive.Separator>
>(({ className, ...props }, ref) => (
	<SelectPrimitive.Separator
		ref={ref}
		className={cn("bg-muted -mx-1 my-1 h-px", className)}
		{...props}
	/>
));
SelectSeparator.displayName = "SelectSeparator";

export {
	Select,
	SelectGroup,
	SelectValue,
	SelectTrigger,
	SelectContent,
	SelectLabel,
	SelectItem,
	SelectSeparator,
	SelectScrollUpButton,
	SelectScrollDownButton,
};
