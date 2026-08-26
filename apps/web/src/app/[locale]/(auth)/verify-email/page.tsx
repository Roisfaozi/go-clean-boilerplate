import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";
import { VerifyEmailStatus } from "~/components/auth/verify-email-status";

export default async function VerifyEmail({
	searchParams,
}: {
	searchParams: Promise<{ token?: string | string[] }>;
}) {
	const { token: rawToken } = await searchParams;
	const token = Array.isArray(rawToken) ? rawToken[0] : rawToken;

	return (
		<AuthLayoutShell
			title="Verify your email"
			description="Confirming your email address keeps your account secure"
			brandingTitle="One Last Security Check."
			brandingDescription="Email verification protects your workspace and ensures important account notifications reach you."
		>
			<VerifyEmailStatus token={token} />
		</AuthLayoutShell>
	);
}
