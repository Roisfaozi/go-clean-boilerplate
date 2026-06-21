import Link from "next/link";
import RegisterForm from "~/components/auth/register-form";
import { AuthLayoutShell } from "~/components/auth/auth-layout-shell";

export default function Register() {
  return (
    <AuthLayoutShell
      title="Create an account"
      description="Enter your details to get started with NexusOS"
      brandingTitle="Start Building Secure Applications in Minutes."
      brandingDescription="NexusOS gives you all the tools you need to build enterprise-grade SaaS with a focus on speed and security."
      features={[
        "Role-Based Access Control (Casbin)",
        "Multi-tenant Architecture",
        "Real-time Audit Logging",
        "Distributed WebSocket Support",
      ]}
      footer={
        <p className="text-muted-foreground text-sm">
          Already have an account?{" "}
          <Link
            href="/login"
            className="text-primary font-medium underline-offset-4 hover:underline"
          >
            Sign in
          </Link>
        </p>
      }
    >
      <RegisterForm />
    </AuthLayoutShell>
  );
}
