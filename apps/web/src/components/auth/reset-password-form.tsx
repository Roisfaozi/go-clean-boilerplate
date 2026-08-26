"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button, Input, Label } from "@casbin/ui";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import type { z } from "zod";
import { Icon } from "~/components/shared/icon";
import { authApi, authTokenSchema, resetPasswordSchema } from "~/lib/api/auth";

type FormData = z.infer<typeof resetPasswordSchema>;

export function ResetPasswordForm({ token }: { token?: string }) {
	const router = useRouter();
	const [isLoading, setIsLoading] = useState(false);
	const parsedToken = authTokenSchema.safeParse(token);
	const resetToken = parsedToken.success ? parsedToken.data : undefined;
	const {
		register,
		handleSubmit,
		formState: { errors },
	} = useForm<FormData>({
		resolver: zodResolver(resetPasswordSchema),
		defaultValues: { password: "", confirmPassword: "" },
	});

	if (!resetToken) {
		return (
			<div className="space-y-4 text-center" role="alert">
				<div className="bg-destructive/10 text-destructive mx-auto flex h-12 w-12 items-center justify-center rounded-full">
					<Icon name="TriangleAlert" className="h-6 w-6" />
				</div>
				<div>
					<h2 className="font-semibold">Invalid reset link</h2>
					<p className="text-muted-foreground mt-1 text-sm">
						This link is missing its reset token. Request a new link to
						continue.
					</p>
				</div>
				<Button render={<Link href="/forgot-password" />}>
					Request New Link
				</Button>
			</div>
		);
	}
	const validResetToken: string = resetToken;

	async function onSubmit(data: FormData) {
		setIsLoading(true);
		try {
			await authApi.resetPassword(validResetToken, data.password);
			toast.success("Password updated", {
				description: "You can now sign in with your new password.",
			});
			router.push("/login");
		} catch (error) {
			toast.error(
				error instanceof Error
					? error.message
					: "Failed to reset your password.",
			);
		} finally {
			setIsLoading(false);
		}
	}

	return (
		<form onSubmit={handleSubmit(onSubmit)} className="grid gap-4">
			<div className="grid gap-2">
				<Label htmlFor="password">New Password</Label>
				<Input
					id="password"
					placeholder="••••••••"
					type="password"
					autoComplete="new-password"
					disabled={isLoading}
					{...register("password")}
				/>
				{errors.password && (
					<p className="text-destructive px-1 text-xs">
						{errors.password.message}
					</p>
				)}
			</div>
			<div className="grid gap-2">
				<Label htmlFor="confirmPassword">Confirm New Password</Label>
				<Input
					id="confirmPassword"
					placeholder="••••••••"
					type="password"
					autoComplete="new-password"
					disabled={isLoading}
					{...register("confirmPassword")}
				/>
				{errors.confirmPassword && (
					<p className="text-destructive px-1 text-xs">
						{errors.confirmPassword.message}
					</p>
				)}
			</div>
			<Button type="submit" className="mt-2 w-full" disabled={isLoading}>
				{isLoading && (
					<Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
				)}
				Update Password
			</Button>
		</form>
	);
}
