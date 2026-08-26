"use client";

import * as React from "react";
import { Tabs } from "@base-ui/react/tabs";

import { cn } from "./lib/utils";

const TabsRoot = Tabs.Root;

const TabsList = React.forwardRef<
	React.ComponentRef<typeof Tabs.List>,
	React.ComponentProps<typeof Tabs.List>
>(({ className, ...props }, ref) => (
	<Tabs.List
		ref={ref}
		className={cn(
			"bg-muted text-muted-foreground inline-flex h-10 items-center justify-center rounded-md p-1",
			className,
		)}
		{...props}
	/>
));
TabsList.displayName = "TabsList";

const TabsTrigger = React.forwardRef<
	React.ComponentRef<typeof Tabs.Tab>,
	React.ComponentProps<typeof Tabs.Tab>
>(({ className, ...props }, ref) => (
	<Tabs.Tab
		ref={ref}
		className={cn(
			"ring-offset-background focus-visible:ring-ring data-[active]:bg-background data-[active]:text-foreground inline-flex items-center justify-center rounded-sm px-3 py-1.5 text-sm font-medium whitespace-nowrap transition-all focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50 data-[active]:shadow-xs",
			className,
		)}
		{...props}
	/>
));
TabsTrigger.displayName = "TabsTab";

const TabsContent = React.forwardRef<
	React.ComponentRef<typeof Tabs.Panel>,
	React.ComponentProps<typeof Tabs.Panel>
>(({ className, ...props }, ref) => (
	<Tabs.Panel
		ref={ref}
		className={cn(
			"ring-offset-background focus-visible:ring-ring mt-2 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none",
			className,
		)}
		{...props}
	/>
));
TabsContent.displayName = "TabsContent";

export { TabsRoot as Tabs, TabsList, TabsTrigger, TabsContent };
