"use client";

import { Toast as ToastPrimitive } from "@base-ui/react/toast";
import * as React from "react";

import { toastManager } from "./hooks/use-toast";
import {
	Toast,
	ToastAction,
	ToastClose,
	ToastDescription,
	ToastProvider,
	ToastTitle,
	ToastViewport,
} from "./toast";

function ToastList() {
	const { toasts } = ToastPrimitive.useToastManager();

	return (
		<>
			{toasts.map((toast) => (
				<Toast key={toast.id} toast={toast}>
					<div className="grid gap-1">
						{toast.title && <ToastTitle>{toast.title}</ToastTitle>}
						{toast.description && (
							<ToastDescription>{toast.description}</ToastDescription>
						)}
					</div>
					{toast.actionProps && <ToastActionFromProps props={toast.actionProps} />}
					<ToastClose />
				</Toast>
			))}
			<ToastViewport />
		</>
	);
}

function ToastActionFromProps({
	props,
}: {
	props: React.ComponentPropsWithoutRef<typeof ToastAction>;
}) {
	return <ToastAction {...props} />;
}

export function Toaster() {
	return (
		<ToastProvider toastManager={toastManager} limit={1}>
			<ToastList />
		</ToastProvider>
	);
}
