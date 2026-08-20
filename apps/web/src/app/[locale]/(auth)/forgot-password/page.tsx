import Link from "next/link";
import { Icon } from "~/components/shared/icon";
import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";
import { ForgotPasswordForm } from "~/components/auth/forgot-password-form";

export default function ForgotPassword() {
	return (
		<AuthLayoutShell
			title="Forgot password?"
			description="Enter your email and we'll send you a link to reset your password"
			brandingTitle="Don't Worry!"
			brandingDescription="It happens to the best of us. We'll have you back in your account in no time."
			footer={
				<Link
					href="/login"
					className="text-muted-foreground hover:text-primary flex items-center gap-2 text-sm font-medium transition-colors"
				>
					<Icon name="ChevronLeft" className="h-4 w-4" />
					Back to login
				</Link>
			}
		>
			<ForgotPasswordForm />
		</AuthLayoutShell>
	);
}
