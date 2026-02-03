import Link from "next/link";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import Icons from "~/components/shared/icons";

export default function ResetPassword() {
  return (
    <div className="flex min-h-screen w-full flex-col lg:flex-row">
      {/* Left Panel: Functional Zone */}
      <div className="flex flex-1 flex-col justify-center px-6 py-12 md:px-12 lg:px-24 xl:px-32">
        <div className="mx-auto w-full max-w-sm lg:mx-0">
          <div className="mb-8">
            <h1 className="text-3xl font-bold tracking-tight">Reset password</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              Enter your new password below
            </p>
          </div>

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
            <Button className="mt-2 w-full">
              Update Password
            </Button>
          </div>
        </div>
      </div>

      {/* Right Panel: Visual/Branding Zone */}
      <div className="relative hidden w-full flex-1 items-center justify-center overflow-hidden bg-primary lg:flex">
        <div className="absolute inset-0 bg-linear-to-br from-indigo-600 to-violet-700 opacity-90" />
        <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/cubes.png')] opacity-20" />
        
        <div className="relative z-10 text-center p-12 text-white">
          <div className="mx-auto mb-8 flex h-24 w-24 items-center justify-center rounded-full bg-white/10 backdrop-blur-md">
            <Icons.settings className="h-12 w-12" />
          </div>
          <h2 className="mb-4 text-3xl font-bold tracking-tight">Almost there!</h2>
          <p className="max-w-md mx-auto text-indigo-100 text-lg">
            Once you update your password, you&apos;ll be able to sign in and access your workspace.
          </p>
        </div>
      </div>
    </div>
  );
}
