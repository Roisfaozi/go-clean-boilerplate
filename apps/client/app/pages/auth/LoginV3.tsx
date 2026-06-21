// V3: Full-screen dark immersive with grid pattern and neon accents
import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { Hexagon, ArrowRight, Lock, User, ChevronRight } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { motion } from "framer-motion";

const loginSchema = z.object({
  username: z.string().trim().min(1, "Username wajib diisi"),
  password: z.string().min(1, "Password wajib diisi"),
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

export default function LoginV3() {
  const [loading, setLoading] = useState(false);
  const [googleLoading, setGoogleLoading] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});
    const result = loginSchema.safeParse({ username, password });
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
        id: "1",
        name: "Admin User",
        email: "admin@nexus.dev",
        username: result.data.username,
      });
      navigate("/");
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleLogin = () => {
    setGoogleLoading(true);
    setTimeout(() => {
      login({
        id: "g1",
        name: "Google User",
        email: "user@gmail.com",
        username: "googleuser",
      });
      navigate("/");
    }, 1500);
  };

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
              The future of
              <br />
              <span className="from-primary via-accent to-secondary bg-gradient-to-r bg-clip-text text-transparent">
                workspace management
              </span>
            </h1>
            <p className="max-w-md text-lg opacity-60">
              Secure, scalable, and beautifully crafted. Your command center
              awaits.
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
              "NexusOS transformed how we manage access control across 200+
              microservices. The RBAC system is incredibly powerful yet
              intuitive."
            </p>
            <div className="mt-4 flex items-center gap-3">
              <div className="bg-primary/30 flex h-8 w-8 items-center justify-center rounded-full text-xs font-bold">
                JD
              </div>
              <div>
                <div className="text-sm font-medium">Jane Doe</div>
                <div className="text-xs opacity-50">CTO, TechCorp</div>
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
          className="w-full max-w-[400px] space-y-8"
        >
          {/* Mobile logo */}
          <div className="mb-2 flex items-center gap-3 lg:hidden">
            <div className="bg-primary flex h-10 w-10 items-center justify-center rounded-xl">
              <Hexagon className="text-primary-foreground h-5 w-5" />
            </div>
            <span className="text-xl font-bold">NexusOS</span>
          </div>

          <div className="space-y-2">
            <h2 className="text-3xl font-bold">Sign in</h2>
            <p className="text-sm opacity-50">
              Access your workspace and projects
            </p>
          </div>

          {/* Google */}
          <NexusButton
            variant="outline"
            className="border-primary-foreground/20 bg-primary-foreground/5 hover:bg-primary-foreground/10 text-primary-foreground h-12 w-full gap-2 rounded-xl"
            onClick={handleGoogleLogin}
            loading={googleLoading}
            disabled={loading}
          >
            <GoogleIcon />
            Continue with Google
          </NexusButton>

          <div className="flex items-center gap-3">
            <div className="bg-primary-foreground/10 h-px flex-1" />
            <span className="text-xs tracking-widest uppercase opacity-40">
              or
            </span>
            <div className="bg-primary-foreground/10 h-px flex-1" />
          </div>

          <form className="space-y-5" onSubmit={handleSubmit}>
            <div className="space-y-1.5">
              <label className="flex items-center gap-2 text-sm font-medium opacity-70">
                <User className="h-3.5 w-3.5" /> Username
              </label>
              <input
                className="border-primary-foreground/15 bg-primary-foreground/5 text-primary-foreground focus:border-primary focus:ring-primary h-12 w-full rounded-xl border px-4 transition-colors placeholder:opacity-30 focus:ring-1 focus:outline-none"
                placeholder="Enter your username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
              />
              {errors.username && (
                <p className="text-danger mt-1 text-xs">{errors.username}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <div className="flex items-center justify-between">
                <label className="flex items-center gap-2 text-sm font-medium opacity-70">
                  <Lock className="h-3.5 w-3.5" /> Password
                </label>
                <Link
                  to="/forgot-password"
                  className="text-primary text-xs hover:underline"
                >
                  Forgot?
                </Link>
              </div>
              <input
                type="password"
                className="border-primary-foreground/15 bg-primary-foreground/5 text-primary-foreground focus:border-primary focus:ring-primary h-12 w-full rounded-xl border px-4 transition-colors placeholder:opacity-30 focus:ring-1 focus:outline-none"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
              />
              {errors.password && (
                <p className="text-danger mt-1 text-xs">{errors.password}</p>
              )}
            </div>
            <NexusButton
              className="bg-primary hover:bg-primary-hover text-primary-foreground h-12 w-full gap-2 rounded-xl"
              loading={loading}
              disabled={googleLoading}
            >
              Sign In <ArrowRight className="h-4 w-4" />
            </NexusButton>
          </form>

          <p className="text-center text-sm opacity-50">
            Don't have an account?{" "}
            <Link
              to="/register"
              className="text-primary inline-flex items-center gap-1 font-medium hover:underline"
            >
              Get started <ChevronRight className="h-3 w-3" />
            </Link>
          </p>
        </motion.div>
      </div>
    </div>
  );
}
