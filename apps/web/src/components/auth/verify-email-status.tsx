"use client";

import { Button } from "@casbin/ui";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { Icon } from "~/components/shared/icon";
import { authApi, authTokenSchema } from "~/lib/api/auth";

type VerificationState = "processing" | "success" | "error" | "invalid";

export function VerifyEmailStatus({ token }: { token?: string }) {
	const parsedToken = authTokenSchema.safeParse(token);
	const verificationToken = parsedToken.success ? parsedToken.data : undefined;
	const [attempt, setAttempt] = useState(0);
	const [state, setState] = useState<VerificationState>(
		verificationToken ? "processing" : "invalid",
	);
	const requestedKeyRef = useRef<string | null>(null);

	useEffect(() => {
		if (!verificationToken) return;
		const requestKey = `${verificationToken}:${attempt}`;
		if (requestedKeyRef.current === requestKey) return;
		requestedKeyRef.current = requestKey;

		setState("processing");
		authApi
			.verifyEmail(verificationToken)
			.then(() => {
				if (requestedKeyRef.current === requestKey) setState("success");
			})
			.catch(() => {
				if (requestedKeyRef.current === requestKey) setState("error");
			});
	}, [verificationToken, attempt]);

	if (state === "processing") {
		return (
			<div className="space-y-4 text-center" role="status">
				<Icon
					name="Loader"
					className="text-primary mx-auto h-10 w-10 animate-spin"
				/>
				<p className="text-muted-foreground text-sm">
					Verifying your email address...
				</p>
			</div>
		);
	}

	if (state === "success") {
		return (
			<div className="space-y-4 text-center" role="status">
				<div className="bg-primary/10 text-primary mx-auto flex h-12 w-12 items-center justify-center rounded-full">
					<Icon name="CircleCheck" className="h-6 w-6" />
				</div>
				<div>
					<h2 className="font-semibold">Email verified</h2>
					<p className="text-muted-foreground mt-1 text-sm">
						Your account is ready. You can now sign in.
					</p>
				</div>
				<Button render={<Link href="/login" />}>Continue to Login</Button>
			</div>
		);
	}

	return (
		<div className="space-y-4 text-center" role="alert">
			<div className="bg-destructive/10 text-destructive mx-auto flex h-12 w-12 items-center justify-center rounded-full">
				<Icon name="TriangleAlert" className="h-6 w-6" />
			</div>
			<div>
				<h2 className="font-semibold">
					{state === "invalid"
						? "Invalid verification link"
						: "Verification failed"}
				</h2>
				<p className="text-muted-foreground mt-1 text-sm">
					{state === "invalid"
						? "This link is missing its verification token."
						: "The link may be invalid or expired. Try it again or request a new link after signing in."}
				</p>
			</div>
			{state === "error" && (
				<Button
					variant="outline"
					onClick={() => setAttempt((value) => value + 1)}
				>
					Try Again
				</Button>
			)}
			<Button
				variant={state === "error" ? "ghost" : "default"}
				render={<Link href="/login" />}
			>
				Go to Login
			</Button>
		</div>
	);
}
