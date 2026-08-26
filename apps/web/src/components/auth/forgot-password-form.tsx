"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button, Input, Label } from "@casbin/ui";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import type { z } from "zod";
import { Icon } from "~/components/shared/icon";
import { authApi, forgotPasswordSchema } from "~/lib/api/auth";

type FormData = z.infer<typeof forgotPasswordSchema>;

export function ForgotPasswordForm() {
	const [isLoading, setIsLoading] = useState(false);
	const [isSubmitted, setIsSubmitted] = useState(false);
	const {
		register,
		handleSubmit,
		formState: { errors },
	} = useForm<FormData>({
		resolver: zodResolver(forgotPasswordSchema),
		defaultValues: { email: "" },
	});

	async function onSubmit(data: FormData) {
		setIsLoading(true);
		try {
			await authApi.forgotPassword(data);
			setIsSubmitted(true);
		} catch (error) {
			toast.error(
				error instanceof Error
					? error.message
					: "Failed to request a password reset.",
			);
		} finally {
			setIsLoading(false);
		}
	}

	if (isSubmitted) {
		return (
			<div className="space-y-4 text-center" role="status">
				<div className="bg-primary/10 text-primary mx-auto flex h-12 w-12 items-center justify-center rounded-full">
					<Icon name="Mail" className="h-6 w-6" />
				</div>
				<div>
					<h2 className="font-semibold">Check your inbox</h2>
					<p className="text-muted-foreground mt-1 text-sm">
						If the email is registered, a password reset link will arrive
						shortly.
					</p>
				</div>
			</div>
		);
	}

	return (
		<form onSubmit={handleSubmit(onSubmit)} className="grid gap-4">
			<div className="grid gap-2">
				<Label htmlFor="email">Email</Label>
				<Input
					id="email"
					placeholder="name@example.com"
					type="email"
					autoCapitalize="none"
					autoComplete="email"
					autoCorrect="off"
					disabled={isLoading}
					{...register("email")}
				/>
				{errors.email && (
					<p className="text-destructive px-1 text-xs">
						{errors.email.message}
					</p>
				)}
			</div>
			<Button type="submit" className="mt-2 w-full" disabled={isLoading}>
				{isLoading && (
					<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
				)}
				Send Reset Link
			</Button>
		</form>
	);
}
