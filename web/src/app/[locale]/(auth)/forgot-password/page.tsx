import Link from "next/link";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { Label } from "~/components/ui/label";
import Icons from "~/components/shared/icons";

export default function ForgotPassword() {
  return (
    <div className="flex min-h-screen w-full flex-col lg:flex-row">
      {/* Left Panel: Functional Zone */}
      <div className="flex flex-1 flex-col justify-center px-6 py-12 md:px-12 lg:px-24 xl:px-32">
        <div className="mx-auto w-full max-w-sm lg:mx-0">
          <Link
            href="/login"
            className="text-muted-foreground hover:text-primary mb-10 flex items-center gap-2 text-sm font-medium transition-colors"
          >
            <Icons.chevronLeft className="h-4 w-4" />
            Back to login
          </Link>

          <div className="mb-8">
            <h1 className="text-3xl font-bold tracking-tight">
              Forgot password?
            </h1>
            <p className="text-muted-foreground mt-2 text-sm">
              Enter your email and we&apos;ll send you a link to reset your
              password
            </p>
          </div>

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
        </div>
      </div>

      {/* Right Panel: Visual/Branding Zone */}
      <div className="bg-primary relative hidden w-full flex-1 items-center justify-center overflow-hidden lg:flex">
        <div className="absolute inset-0 bg-linear-to-br from-indigo-600 to-violet-700 opacity-90" />
        <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/cubes.png')] opacity-20" />

        <div className="relative z-10 p-12 text-center text-white">
          <div className="mx-auto mb-8 flex h-24 w-24 items-center justify-center rounded-full bg-white/10 backdrop-blur-md">
            <Icons.help className="h-12 w-12" />
          </div>
          <h2 className="mb-4 text-3xl font-bold tracking-tight">
            Don&apos;t Worry!
          </h2>
          <p className="mx-auto max-w-md text-lg text-indigo-100">
            It happens to the best of us. We&apos;ll have you back in your
            account in no time.
          </p>
        </div>
      </div>
    </div>
  );
}
