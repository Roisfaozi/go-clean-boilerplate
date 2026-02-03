import Link from 'next/link';
import AuthForm from "~/components/auth/login-form";
import Icons from "~/components/shared/icons";

export default function Login() {
  return (
    <div className="flex min-h-screen w-full flex-col lg:flex-row">
      {/* Left Panel: Functional Zone */}
      <div className="flex flex-1 flex-col justify-center px-6 py-12 md:px-12 lg:px-24 xl:px-32">
        <div className="mx-auto w-full max-w-sm lg:mx-0">
          <Link href="/" className="mb-10 flex items-center gap-2">
            <Icons.logo className="h-10 w-10 text-primary" />
            <span className="text-2xl font-bold tracking-tighter">NexusOS</span>
          </Link>

          <div className="mb-8">
            <h1 className="text-3xl font-bold tracking-tight">Welcome back</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              Enter your credentials to access your account
            </p>
          </div>

          <AuthForm />

          <p className="mt-8 text-center text-sm text-muted-foreground lg:text-left">
            Don&apos;t have an account?{" "}
            <Link
              href="/register"
              className="font-medium text-primary underline-offset-4 hover:underline"
            >
              Create account
            </Link>
          </p>
        </div>
      </div>

      {/* Right Panel: Visual/Branding Zone */}
      <div className="relative hidden w-full flex-1 items-center justify-center overflow-hidden bg-primary lg:flex">
        {/* Background Decorative Elements */}
        <div className="absolute inset-0 bg-linear-to-br from-indigo-600 to-violet-700 opacity-90" />
        <div className="absolute inset-0 bg-[url('https://www.transparenttextures.com/patterns/cubes.png')] opacity-20" />

        <div className="relative z-10 p-12 text-white">
          <div className="max-w-md">
            <div className="mb-8 flex h-12 w-12 items-center justify-center rounded-xl bg-white/20 backdrop-blur-md">
              <Icons.logo className="h-8 w-8" />
            </div>
            <h2 className="mb-6 text-4xl font-bold leading-tight tracking-tight">
              Enterprise-grade RBAC and Real-time Monitoring.
            </h2>
            <p className="mb-10 text-lg text-indigo-100">
              NexusOS provides the most robust boilerplate for building complex,
              secure, and scalable multi-tenant applications.
            </p>

            <div className="rounded-2xl bg-white/10 p-6 backdrop-blur-lg">
              <p className="italic text-indigo-50">
                &quot;NexusOS has significantly reduced our development time for
                internal tools. The Casbin integration is seamless and powerful.&quot;
              </p>
              <div className="mt-4 flex items-center gap-3">
                <div className="h-10 w-10 rounded-full bg-indigo-300/50" />
                <div>
                  <p className="text-sm font-semibold">Sarah Jenkins</p>
                  <p className="text-xs text-indigo-200">CTO at TechFlow</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Animated Orbs */}
        <div className="absolute -bottom-24 -right-24 h-96 w-96 rounded-full bg-violet-500/20 blur-3xl animate-pulse" />
        <div className="absolute -left-24 -top-24 h-96 w-96 rounded-full bg-indigo-400/20 blur-3xl animate-pulse" />
      </div>
    </div>
  );
}
