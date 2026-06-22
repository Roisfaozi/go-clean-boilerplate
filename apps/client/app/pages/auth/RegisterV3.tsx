// V3: Dark immersive register — matches LoginV3 style
import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import {
  Hexagon,
  ArrowRight,
  User,
  Mail,
  Lock,
  KeyRound,
  ChevronRight,
} from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
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

export default function RegisterV3() {
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

  const inputClass =
    "w-full h-12 rounded-xl border border-primary-foreground/15 bg-primary-foreground/5 px-4 text-primary-foreground placeholder:opacity-30 focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-colors";

  return (
    <div className="bg-foreground text-primary-foreground relative flex min-h-screen">
      {/* Grid pattern */}
      <div
        className="absolute inset-0 opacity-[0.04]"
        style={{
          backgroundImage:
            "linear-gradient(hsl(var(--primary-foreground)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--primary-foreground)) 1px, transparent 1px)",
          backgroundSize: "60px 60px",
        }}
      />

      {/* Glow effects */}
      <div className="bg-primary/20 absolute top-0 left-1/2 h-[400px] w-[600px] -translate-x-1/2 rounded-full blur-[120px]" />
      <div className="bg-accent/15 absolute right-0 bottom-0 h-[300px] w-[400px] rounded-full blur-[100px]" />

      {/* Left: Branding */}
      <div className="relative z-10 hidden flex-col justify-center p-16 lg:flex lg:w-1/2">
        <motion.div
          initial={{ opacity: 0, x: -30 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.7 }}
          className="max-w-lg space-y-10"
        >
          <div className="flex items-center gap-3">
            <div className="bg-primary flex h-12 w-12 items-center justify-center rounded-xl">
              <Hexagon className="text-primary-foreground h-6 w-6" />
            </div>
            <span className="text-2xl font-bold tracking-tight">NexusOS</span>
          </div>

          <div className="space-y-4">
            <h1 className="text-5xl leading-[1.1] font-bold tracking-tight">
              Join the future of
              <br />
              <span className="from-primary via-accent to-secondary bg-gradient-to-r bg-clip-text text-transparent">
                workspace management
              </span>
            </h1>
            <p className="max-w-md text-lg opacity-60">
              Create your free account and start managing teams, roles, and
              projects with ease.
            </p>
          </div>

          {/* Stats row */}
          <div className="flex gap-8">
            {[
              { value: "99.9%", label: "Uptime" },
              { value: "50K+", label: "Teams" },
              { value: "<100ms", label: "Latency" },
            ].map(({ value, label }) => (
              <div key={label}>
                <div className="text-primary text-2xl font-bold">{value}</div>
                <div className="text-sm opacity-50">{label}</div>
              </div>
            ))}
          </div>

          {/* Testimonial */}
          <div className="border-primary-foreground/10 bg-primary-foreground/5 rounded-2xl border p-6 backdrop-blur-sm">
            <p className="text-sm leading-relaxed italic opacity-70">
              "Setting up our entire team on NexusOS took less than 10 minutes.
              The onboarding experience is seamless."
            </p>
            <div className="mt-4 flex items-center gap-3">
              <div className="bg-primary/30 flex h-8 w-8 items-center justify-center rounded-full text-xs font-bold">
                SK
              </div>
              <div>
                <div className="text-sm font-medium">Sarah Kim</div>
                <div className="text-xs opacity-50">
                  VP Engineering, StartupXYZ
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Right: Form */}
      <div className="relative z-10 flex flex-1 items-center justify-center p-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.2 }}
          className="w-full max-w-[420px] space-y-6"
        >
          {/* Mobile logo */}
          <div className="mb-2 flex items-center gap-3 lg:hidden">
            <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-xl">
              <Hexagon className="text-primary-foreground h-5 w-5" />
            </div>
            <span className="text-xl font-bold">NexusOS</span>
          </div>

          <div className="space-y-2">
            <h2 className="text-3xl font-bold">Create account</h2>
            <p className="text-sm opacity-50">
              Free forever. No credit card required.
            </p>
          </div>

          <NexusButton
            variant="outline"
            className="border-primary-foreground/20 bg-primary-foreground/5 hover:bg-primary-foreground/10 text-primary-foreground h-12 w-full gap-2 rounded-xl"
            onClick={handleGoogleSignup}
            disabled={loading}
          >
            <GoogleIcon /> Sign up with Google
          </NexusButton>

          <div className="flex items-center gap-3">
            <div className="bg-primary-foreground/10 h-px flex-1" />
            <span className="text-xs tracking-widest uppercase opacity-40">
              or
            </span>
            <div className="bg-primary-foreground/10 h-px flex-1" />
          </div>

          <form className="space-y-4" onSubmit={handleSubmit}>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <label className="flex items-center gap-2 text-sm font-medium opacity-70">
                  <User className="h-3.5 w-3.5" /> Full Name
                </label>
                <input
                  className={inputClass}
                  placeholder="John Doe"
                  value={form.name}
                  onChange={update("name")}
                  disabled={loading}
                />
                {errors.name && (
                  <p className="text-destructive mt-1 text-xs">{errors.name}</p>
                )}
              </div>
              <div className="space-y-1.5">
                <label className="flex items-center gap-2 text-sm font-medium opacity-70">
                  <User className="h-3.5 w-3.5" /> Username
                </label>
                <input
                  className={inputClass}
                  placeholder="johndoe"
                  value={form.username}
                  onChange={update("username")}
                  disabled={loading}
                />
                {errors.username && (
                  <p className="text-destructive mt-1 text-xs">
                    {errors.username}
                  </p>
                )}
              </div>
            </div>
            <div className="space-y-1.5">
              <label className="flex items-center gap-2 text-sm font-medium opacity-70">
                <Mail className="h-3.5 w-3.5" /> Email
              </label>
              <input
                className={inputClass}
                type="email"
                placeholder="john@example.com"
                value={form.email}
                onChange={update("email")}
                disabled={loading}
              />
              {errors.email && (
                <p className="text-destructive mt-1 text-xs">{errors.email}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <label className="flex items-center gap-2 text-sm font-medium opacity-70">
                <Lock className="h-3.5 w-3.5" /> Password
              </label>
              <input
                className={inputClass}
                type="password"
                placeholder="Min. 8 characters"
                value={form.password}
                onChange={update("password")}
                disabled={loading}
              />
              {form.password.length > 0 && (
                <div className="mt-2 flex items-center gap-2">
                  <div className="flex flex-1 gap-1">
                    {[1, 2, 3].map((level) => (
                      <div
                        key={level}
                        className={`h-1 flex-1 rounded-full transition-colors ${strength >= level ? strengthColor : "bg-primary-foreground/10"}`}
                      />
                    ))}
                  </div>
                  <span className="text-xs opacity-50">{strengthLabel}</span>
                </div>
              )}
              {errors.password && (
                <p className="text-destructive mt-1 text-xs">
                  {errors.password}
                </p>
              )}
            </div>
            <div className="space-y-1.5">
              <label className="flex items-center gap-2 text-sm font-medium opacity-70">
                <KeyRound className="h-3.5 w-3.5" /> Confirm Password
              </label>
              <input
                className={inputClass}
                type="password"
                placeholder="Confirm password"
                value={form.confirmPassword}
                onChange={update("confirmPassword")}
                disabled={loading}
              />
              {errors.confirmPassword && (
                <p className="text-destructive mt-1 text-xs">
                  {errors.confirmPassword}
                </p>
              )}
            </div>
            <NexusButton
              className="bg-primary hover:bg-primary-hover text-primary-foreground h-12 w-full gap-2 rounded-xl"
              loading={loading}
            >
              Create Account <ArrowRight className="h-4 w-4" />
            </NexusButton>
          </form>

          <p className="text-center text-sm opacity-50">
            Already have an account?{" "}
            <Link
              to="/login"
              className="text-primary inline-flex items-center gap-1 font-medium hover:underline"
            >
              Sign in <ChevronRight className="h-3 w-3" />
            </Link>
          </p>
          <p className="text-center text-[11px] opacity-30">
            By creating an account you agree to our Terms of Service and Privacy
            Policy
          </p>
        </motion.div>
      </div>
    </div>
  );
}
