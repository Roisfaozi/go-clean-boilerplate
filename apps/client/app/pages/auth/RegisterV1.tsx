// V1: Split-screen register — matches LoginV1 style
import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { Hexagon, Shield, Zap, Globe, ArrowRight } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { Separator } from "@casbin/ui";
import { motion } from "framer-motion";

const registerSchema = z
  .object({
    name: z.string().trim().min(1, "Name is required").max(100),
    email: z.string().trim().email("Invalid email").max(255),
    username: z.string().trim().min(3, "Min 3 characters").max(50),
    password: z.string().min(8, "Min 8 characters"),
    confirmPassword: z.string().min(1, "Please confirm password"),
  })
  .refine((d) => d.password === d.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

const GoogleIcon = () => (
  <svg className="h-4 w-4" viewBox="0 0 24 24">
    <path
      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
      fill="#4285F4"
    />
    <path
      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
      fill="#34A853"
    />
    <path
      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18A10.96 10.96 0 0 0 1 12c0 1.77.42 3.45 1.18 4.93l3.66-2.84z"
      fill="#FBBC05"
    />
    <path
      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
      fill="#EA4335"
    />
  </svg>
);

export default function RegisterV1() {
  const [loading, setLoading] = useState(false);
  const [form, setForm] = useState({
    name: "",
    email: "",
    username: "",
    password: "",
    confirmPassword: "",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);

  const update = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});
    const result = registerSchema.safeParse(form);
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      result.error.errors.forEach((err) => {
        if (err.path[0]) fieldErrors[String(err.path[0])] = err.message;
      });
      setErrors(fieldErrors);
      return;
    }
    setLoading(true);
    try {
      login({
        id: "new1",
        name: result.data.name,
        email: result.data.email,
        username: result.data.username,
      });
      navigate("/");
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleSignup = () => {
    login({
      id: "g1",
      name: "Google User",
      email: "user@gmail.com",
      username: "googleuser",
    });
    navigate("/");
  };

  const strength =
    form.password.length === 0
      ? 0
      : form.password.length < 8
        ? 1
        : form.password.length < 12
          ? 2
          : 3;
  const strengthLabel = ["", "Weak", "Good", "Strong"][strength];
  const strengthColor = ["", "bg-destructive", "bg-warning", "bg-success"][
    strength
  ];

  return (
    <div className="flex min-h-screen">
      {/* Left: Branding */}
      <div className="from-primary via-accent to-primary relative hidden overflow-hidden bg-gradient-to-br lg:flex lg:w-1/2">
        <div
          className="absolute inset-0 opacity-10"
          style={{
            backgroundImage:
              "radial-gradient(circle at 1px 1px, hsl(var(--primary-foreground)) 1px, transparent 0)",
            backgroundSize: "32px 32px",
          }}
        />
        <motion.div
          animate={{ y: [0, -20, 0], rotate: [0, 5, 0] }}
          transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }}
          className="border-primary-foreground/20 bg-primary-foreground/5 absolute top-20 right-20 h-32 w-32 rounded-2xl border backdrop-blur-sm"
        />
        <motion.div
          animate={{ y: [0, 15, 0], rotate: [0, -3, 0] }}
          transition={{
            duration: 8,
            repeat: Infinity,
            ease: "easeInOut",
            delay: 1,
          }}
          className="border-primary-foreground/20 bg-primary-foreground/5 absolute bottom-32 left-16 h-24 w-24 rounded-full border backdrop-blur-sm"
        />
        <motion.div
          animate={{ y: [0, -10, 0] }}
          transition={{
            duration: 5,
            repeat: Infinity,
            ease: "easeInOut",
            delay: 2,
          }}
          className="border-primary-foreground/15 bg-primary-foreground/5 absolute top-1/2 right-1/3 h-16 w-16 rotate-45 rounded-lg border backdrop-blur-sm"
        />
        <div className="text-primary-foreground relative z-10 flex w-full flex-col justify-between p-12">
          <div className="flex items-center gap-3">
            <Hexagon className="h-8 w-8" />
            <span className="text-xl font-bold tracking-tight">NexusOS</span>
          </div>
          <div className="max-w-md space-y-8">
            <motion.h2
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6 }}
              className="text-4xl leading-tight font-bold"
            >
              Start building
              <br />
              <span className="opacity-80">something great.</span>
            </motion.h2>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.2 }}
              className="space-y-4"
            >
              {[
                { icon: Shield, text: "Enterprise-grade security & RBAC" },
                { icon: Zap, text: "Real-time analytics & insights" },
                { icon: Globe, text: "Multi-organization workspace" },
              ].map(({ icon: Icon, text }) => (
                <div
                  key={text}
                  className="flex items-center gap-3 text-sm opacity-90"
                >
                  <div className="bg-primary-foreground/10 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg">
                    <Icon className="h-4 w-4" />
                  </div>
                  <span>{text}</span>
                </div>
              ))}
            </motion.div>
          </div>
          <p className="text-xs opacity-60">
            © 2026 NexusOS. All rights reserved.
          </p>
        </div>
      </div>

      {/* Right: Form */}
      <div className="bg-background flex flex-1 items-center justify-center p-6 sm:p-12">
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5 }}
          className="w-full max-w-[440px] space-y-6"
        >
          <div className="mb-2 flex items-center gap-3 lg:hidden">
            <Hexagon className="text-primary h-8 w-8" />
            <span className="text-foreground text-xl font-bold">NexusOS</span>
          </div>
          <div className="space-y-1">
            <h1 className="text-foreground text-3xl font-bold tracking-tight">
              Create account
            </h1>
            <p className="text-muted-foreground">
              Free forever. No credit card required.
            </p>
          </div>

          <NexusButton
            variant="outline"
            className="h-11 w-full gap-2"
            onClick={handleGoogleSignup}
            disabled={loading}
          >
            <GoogleIcon /> Sign up with Google
          </NexusButton>

          <div className="flex items-center gap-3">
            <Separator className="flex-1" />
            <span className="text-muted-foreground text-xs tracking-wider uppercase">
              or
            </span>
            <Separator className="flex-1" />
          </div>

          <form className="space-y-4" onSubmit={handleSubmit}>
            <div className="grid grid-cols-2 gap-4">
              <FormGroup label="Full Name" required error={errors.name}>
                <NexusInput
                  placeholder="John Doe"
                  value={form.name}
                  onChange={update("name")}
                  disabled={loading}
                  className="h-11"
                />
              </FormGroup>
              <FormGroup label="Username" required error={errors.username}>
                <NexusInput
                  placeholder="johndoe"
                  value={form.username}
                  onChange={update("username")}
                  disabled={loading}
                  className="h-11"
                />
              </FormGroup>
            </div>
            <FormGroup label="Email" required error={errors.email}>
              <NexusInput
                type="email"
                placeholder="john@example.com"
                value={form.email}
                onChange={update("email")}
                disabled={loading}
                className="h-11"
              />
            </FormGroup>
            <FormGroup label="Password" required error={errors.password}>
              <NexusInput
                type="password"
                placeholder="Min. 8 characters"
                value={form.password}
                onChange={update("password")}
                disabled={loading}
                className="h-11"
              />
              {form.password.length > 0 && (
                <div className="mt-2 flex items-center gap-2">
                  <div className="flex flex-1 gap-1">
                    {[1, 2, 3].map((level) => (
                      <div
                        key={level}
                        className={`h-1 flex-1 rounded-full transition-colors ${strength >= level ? strengthColor : "bg-muted"}`}
                      />
                    ))}
                  </div>
                  <span className="text-muted-foreground text-xs">
                    {strengthLabel}
                  </span>
                </div>
              )}
            </FormGroup>
            <FormGroup
              label="Confirm Password"
              required
              error={errors.confirmPassword}
            >
              <NexusInput
                type="password"
                placeholder="Confirm password"
                value={form.confirmPassword}
                onChange={update("confirmPassword")}
                disabled={loading}
                className="h-11"
              />
            </FormGroup>
            <NexusButton className="h-11 w-full gap-2" loading={loading}>
              Create Account <ArrowRight className="h-4 w-4" />
            </NexusButton>
          </form>

          <p className="text-muted-foreground text-center text-sm">
            Already have an account?{" "}
            <Link
              to="/login"
              className="text-primary font-medium hover:underline"
            >
              Sign in
            </Link>
          </p>
          <p className="text-muted-foreground/60 text-center text-[11px]">
            By creating an account you agree to our Terms of Service and Privacy
            Policy
          </p>
        </motion.div>
      </div>
    </div>
  );
}
