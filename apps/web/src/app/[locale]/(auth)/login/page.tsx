import Link from "next/link";
import AuthForm from "~/components/auth/login-form";
import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";

export default function Login() {
  return (
    <AuthLayoutShell
      title="Welcome back"
      description="Enter your credentials to access your account"
      brandingTitle="Enterprise-grade RBAC and Real-time Monitoring."
      brandingDescription="NexusOS provides the most robust boilerplate for building complex, secure, and scalable multi-tenant applications."
      testimonial={{
        quote:
          "NexusOS has significantly reduced our development time for internal tools. The Casbin integration is seamless and powerful.",
        author: "Sarah Jenkins",
        role: "CTO at TechFlow",
      }}
      footer={
        <p className="text-muted-foreground text-sm">
          Don&apos;t have an account?{" "}
          <Link
            href="/register"
            className="text-primary font-medium underline-offset-4 hover:underline"
          >
            Create account
          </Link>
        </p>
      }
    >
      <AuthForm />
    </AuthLayoutShell>
  );
}
