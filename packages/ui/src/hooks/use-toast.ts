"use client";

import { Toast as ToastPrimitive } from "@base-ui/react/toast";
import * as React from "react";

import type { ToastActionElement, ToastProps } from "../toast";

export const toastManager = ToastPrimitive.createToastManager();

type ToastVariant = NonNullable<ToastProps["variant"]>;

type ToasterToast = Omit<ToastProps, "toast"> & {
	id: string;
	title?: React.ReactNode;
	description?: React.ReactNode;
	action?: ToastActionElement;
	variant?: ToastVariant;
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
};

type Toast = Omit<ToasterToast, "id">;

interface ToastFunction {
	(
		props: Toast,
	): {
		id: string;
		dismiss: () => void;
		update: (props: ToasterToast) => void;
	};
	success: (title: string, props?: Toast) => void;
	error: (title: string, props?: Toast) => void;
	warning: (title: string, props?: Toast) => void;
}

const toast = (({ variant, action, onOpenChange, open, ...props }: Toast) => {
	const id = toastManager.add({
		...props,
		type: variant,
		actionProps: action?.props as
			| React.ComponentPropsWithoutRef<"button">
			| undefined,
		onClose: () => {
			onOpenChange?.(false);
		},
	});

	const dismiss = () => toastManager.close(id);
	const update = (nextProps: ToasterToast) => {
		const { variant: nextVariant, action: nextAction, ...rest } = nextProps;
		toastManager.update(id, {
			...rest,
			type: nextVariant,
			actionProps: nextAction?.props as
				| React.ComponentPropsWithoutRef<"button">
				| undefined,
		});
	};

	return {
		id,
		dismiss,
		update,
	};
}) as ToastFunction;

toast.success = (title: string, props?: Toast) =>
	toast({ ...props, title, variant: "default" });
toast.error = (title: string, props?: Toast) =>
	toast({ ...props, title, variant: "destructive" });
toast.warning = (title: string, props?: Toast) =>
	toast({ ...props, title, variant: "default" });

function useToast() {
	return {
		toasts: [] as ToasterToast[],
		toast,
		dismiss: (toastId?: string) => toastManager.close(toastId),
	};
}

export { useToast, toast };
