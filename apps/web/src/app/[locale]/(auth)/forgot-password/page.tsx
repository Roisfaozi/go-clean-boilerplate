import Link from "next/link";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { Icon } from "~/components/shared/icon";
import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";

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
      <div className="grid gap-4">
        <div className="grid gap-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            placeholder="name@example.com"
            type="email"
            autoCapitalize="none"
            autoComplete="email"
            autoCorrect="off"
          />
        </div>
        <Button className="mt-2 w-full">Send Reset Link</Button>
      </div>
    </AuthLayoutShell>
  );
}
