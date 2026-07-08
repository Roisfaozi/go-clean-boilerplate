"use client";

import { Toast as ToastPrimitive } from "@base-ui/react/toast";
import { cva, type VariantProps } from "class-variance-authority";
import { X } from "lucide-react";
import * as React from "react";

import { cn } from "./lib/utils";

const ToastProvider = ToastPrimitive.Provider;

const ToastViewport = React.forwardRef<
	React.ElementRef<typeof ToastPrimitive.Viewport>,
	React.ComponentProps<typeof ToastPrimitive.Viewport>
>(({ className, ...props }, ref) => (
	<ToastPrimitive.Viewport
		ref={ref}
		className={cn(
			"fixed z-100 flex max-h-screen w-full flex-col-reverse p-4 md:max-w-[420px]",
			"top-0 right-0 sm:flex-col",
			"[data-density=compact]:top-auto [data-density=compact]:right-0 [data-density=compact]:bottom-0",
			className,
		)}
		{...props}
	/>
));
ToastViewport.displayName = "ToastViewport";

const toastVariants = cva(
	"group pointer-events-auto relative flex w-full items-center justify-between space-x-4 overflow-hidden border p-6 pr-8 shadow-lg transition-all data-[swipe=cancel]:translate-x-0 data-[swipe=end]:translate-x-(--radix-toast-swipe-end-x) data-[swipe=move]:translate-x-(--radix-toast-swipe-move-x) data-[swipe=move]:transition-none data-open:animate-in data-closed:animate-out data-[swipe=end]:animate-out data-closed:fade-out-80 data-closed:slide-out-to-right-full data-open:slide-in-from-top-full sm:data-open:slide-in-from-bottom-full rounded-[var(--radius-lg)]",
	{
		variants: {
			variant: {
				default: "border bg-background text-foreground",
				destructive:
					"destructive group border-destructive bg-destructive text-destructive-foreground",
				success:
					"border-emerald-500 bg-emerald-50 text-emerald-900 dark:bg-emerald-900 dark:text-emerald-50",
				warning:
					"border-amber-500 bg-amber-50 text-amber-900 dark:bg-amber-900 dark:text-amber-50",
				info: "border-blue-500 bg-blue-50 text-blue-900 dark:bg-blue-900 dark:text-blue-50",
				ai: "border-violet-500 bg-violet-50 text-violet-900 dark:bg-violet-900 dark:text-violet-50 animate-pulse",
			},
		},
		defaultVariants: {
			variant: "default",
		},
	},
);

const Toast = React.forwardRef<
	React.ElementRef<typeof ToastPrimitive.Root>,
	React.ComponentProps<typeof ToastPrimitive.Root> &
		VariantProps<typeof toastVariants>
>(({ className, variant, toast, ...props }, ref) => {
	const resolvedVariant = variant ?? (toast.type as VariantProps<typeof toastVariants>["variant"]) ?? "default";
	return (
		<ToastPrimitive.Root
			ref={ref}
			toast={toast}
			className={cn(toastVariants({ variant: resolvedVariant }), className)}
			{...props}
		/>
	);
});
Toast.displayName = "Toast";

const ToastAction = React.forwardRef<
	React.ElementRef<typeof ToastPrimitive.Action>,
	React.ComponentProps<typeof ToastPrimitive.Action>
>(({ className, ...props }, ref) => (
	<ToastPrimitive.Action
		ref={ref}
		className={cn(
			"ring-offset-background hover:bg-secondary focus:ring-ring group-[.destructive]:border-muted/40 hover:group-[.destructive]:border-destructive/30 hover:group-[.destructive]:bg-destructive hover:group-[.destructive]:text-destructive-foreground focus:group-[.destructive]:ring-destructive inline-flex h-8 shrink-0 items-center justify-center rounded-md border bg-transparent px-3 text-sm font-medium transition-colors focus:ring-2 focus:ring-offset-2 focus:outline-hidden disabled:pointer-events-none disabled:opacity-50",
			className,
		)}
		{...props}
	/>
));
ToastAction.displayName = "ToastAction";

const ToastClose = React.forwardRef<
	React.ElementRef<typeof ToastPrimitive.Close>,
	React.ComponentProps<typeof ToastPrimitive.Close>
>(({ className, ...props }, ref) => (
	<ToastPrimitive.Close
		ref={ref}
		className={cn(
			"text-foreground/50 hover:text-foreground absolute top-2 right-2 rounded-md p-1 opacity-0 transition-opacity group-hover:opacity-100 group-[.destructive]:text-red-300 hover:group-[.destructive]:text-red-50 focus:opacity-100 focus:ring-2 focus:outline-hidden focus:group-[.destructive]:ring-red-400 focus:group-[.destructive]:ring-offset-red-600",
			className,
		)}
		toast-close=""
		{...props}
	>
		<X className="h-4 w-4" />
	</ToastPrimitive.Close>
));
ToastClose.displayName = "ToastClose";

const ToastTitle = React.forwardRef<
	React.ElementRef<typeof ToastPrimitive.Title>,
	React.ComponentProps<typeof ToastPrimitive.Title>
>(({ className, ...props }, ref) => (
	<ToastPrimitive.Title
		ref={ref}
		className={cn("text-sm font-semibold", className)}
		{...props}
	/>
));
ToastTitle.displayName = "ToastTitle";

const ToastDescription = React.forwardRef<
	React.ElementRef<typeof ToastPrimitive.Description>,
	React.ComponentProps<typeof ToastPrimitive.Description>
>(({ className, ...props }, ref) => (
	<ToastPrimitive.Description
		ref={ref}
		className={cn("text-sm opacity-90", className)}
		{...props}
	/>
));
ToastDescription.displayName = "ToastDescription";

type ToastProps = React.ComponentProps<typeof Toast>;
type ToastActionElement = React.ReactElement<
	React.ComponentPropsWithoutRef<typeof ToastAction>,
	typeof ToastAction
>;

export {
	type ToastProps,
	type ToastActionElement,
	ToastProvider,
	ToastViewport,
	Toast,
	ToastTitle,
	ToastDescription,
	ToastClose,
	ToastAction,
};
