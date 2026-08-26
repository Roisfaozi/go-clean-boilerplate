"use client";

import * as React from "react";
import { Accordion } from "@base-ui/react/accordion";
import { ChevronDown } from "lucide-react";

import { cn } from "./lib/utils";

const AccordionRoot = Accordion.Root;

const AccordionItem = React.forwardRef<
	React.ComponentRef<typeof Accordion.Item>,
	React.ComponentProps<typeof Accordion.Item>
>(({ className, ...props }, ref) => (
	<Accordion.Item ref={ref} className={cn("border-b", className)} {...props} />
));
AccordionItem.displayName = "AccordionItem";

const AccordionTrigger = React.forwardRef<
	React.ComponentRef<typeof Accordion.Trigger>,
	React.ComponentProps<typeof Accordion.Trigger>
>(({ className, children, ...props }, ref) => (
	<Accordion.Header className="flex">
		<Accordion.Trigger
			ref={ref}
			className={cn(
				"flex flex-1 items-center justify-between py-4 font-medium transition-all hover:underline [&[data-open]>svg]:rotate-180 aria-disabled:cursor-not-allowed aria-disabled:opacity-50",
				className,
			)}
			{...props}
		>
			{children}
			<ChevronDown className="h-4 w-4 shrink-0 transition-transform duration-200" />
		</Accordion.Trigger>
	</Accordion.Header>
));
AccordionTrigger.displayName = "AccordionTrigger";

const AccordionContent = React.forwardRef<
	React.ComponentRef<typeof Accordion.Panel>,
	React.ComponentProps<typeof Accordion.Panel>
>(({ className, children, ...props }, ref) => (
	<Accordion.Panel
		ref={ref}
		className="overflow-hidden text-sm transition-all data-starting-style:h-0 data-ending-style:h-0 data-starting-style:opacity-0 data-ending-style:opacity-0"
		{...props}
	>
		<div className={cn("pt-0 pb-4", className)}>{children}</div>
	</Accordion.Panel>
));
AccordionContent.displayName = "AccordionContent";

export {
	AccordionRoot as Accordion,
	AccordionItem,
	AccordionTrigger,
	AccordionContent,
};
