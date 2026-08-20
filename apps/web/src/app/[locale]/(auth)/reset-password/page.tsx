import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";
import { ResetPasswordForm } from "~/components/auth/reset-password-form";

export default async function ResetPassword({
	searchParams,
}: {
	searchParams: Promise<{ token?: string | string[] }>;
}) {
	const { token: rawToken } = await searchParams;
	const token = Array.isArray(rawToken) ? rawToken[0] : rawToken;

	return (
		<AuthLayoutShell
			title="Reset password"
			description="Enter your new password below"
			brandingTitle="Almost there!"
			brandingDescription="Once you update your password, you'll be able to sign in and access your workspace."
		>
			<ResetPasswordForm token={token} />
		</AuthLayoutShell>
	);
}
