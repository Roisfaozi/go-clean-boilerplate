import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";

export default function ResetPassword() {
  return (
    <AuthLayoutShell
      title="Reset password"
      description="Enter your new password below"
      brandingTitle="Almost there!"
      brandingDescription="Once you update your password, you'll be able to sign in and access your workspace."
    >
      <div className="grid gap-4">
        <div className="grid gap-2">
          <Label htmlFor="password">New Password</Label>
          <Input
            id="password"
            placeholder="••••••••"
            type="password"
            autoComplete="new-password"
          />
        </div>
        <div className="grid gap-2">
          <Label htmlFor="confirmPassword">Confirm New Password</Label>
          <Input
            id="confirmPassword"
            placeholder="••••••••"
            type="password"
            autoComplete="new-password"
          />
        </div>
        <Button className="mt-2 w-full">Update Password</Button>
      </div>
    </AuthLayoutShell>
  );
}
