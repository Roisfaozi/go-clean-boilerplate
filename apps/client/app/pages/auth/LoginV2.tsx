// V2: Centered glassmorphism card on animated gradient background
import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { Hexagon, Sparkles } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { Separator } from "@casbin/ui";
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

export default function LoginV2() {
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
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden">
      {/* Animated gradient background */}
      <div className="from-primary/20 via-accent/15 to-secondary/20 absolute inset-0 bg-gradient-to-br" />
      <motion.div
        animate={{ scale: [1, 1.2, 1], rotate: [0, 90, 180] }}
        transition={{ duration: 20, repeat: Infinity, ease: "linear" }}
        className="absolute -top-1/2 -left-1/2 h-[200%] w-[200%] opacity-30"
        style={{
          background:
            "conic-gradient(from 0deg, hsl(var(--primary)/0.3), hsl(var(--accent)/0.2), hsl(var(--secondary)/0.3), hsl(var(--primary)/0.3))",
        }}
      />

      {/* Floating orbs */}
      <motion.div
        animate={{ y: [0, -30, 0], x: [0, 20, 0] }}
        transition={{ duration: 7, repeat: Infinity, ease: "easeInOut" }}
        className="bg-primary/10 absolute top-1/4 left-1/4 h-64 w-64 rounded-full blur-3xl"
      />
      <motion.div
        animate={{ y: [0, 20, 0], x: [0, -15, 0] }}
        transition={{
          duration: 9,
          repeat: Infinity,
          ease: "easeInOut",
          delay: 2,
        }}
        className="bg-accent/10 absolute right-1/4 bottom-1/4 h-48 w-48 rounded-full blur-3xl"
      />

      {/* Glassmorphism card */}
      <motion.div
        initial={{ opacity: 0, y: 30, scale: 0.95 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
        className="relative z-10 mx-4 w-full max-w-md"
      >
        <div className="border-border/50 bg-card/80 space-y-7 rounded-3xl border p-8 shadow-2xl backdrop-blur-xl">
          {/* Header */}
          <div className="space-y-3 text-center">
            <motion.div
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ type: "spring", delay: 0.2 }}
              className="from-primary to-accent mx-auto inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br"
            >
              <Hexagon className="text-primary-foreground h-8 w-8" />
            </motion.div>
            <h1 className="text-foreground text-2xl font-bold">
              Welcome to NexusOS
            </h1>
            <p className="text-muted-foreground text-sm">
              Sign in to continue to your dashboard
            </p>
          </div>

          {/* Google */}
          <NexusButton
            variant="outline"
            className="h-11 w-full gap-2 rounded-xl"
            onClick={handleGoogleLogin}
            loading={googleLoading}
            disabled={loading}
          >
            <GoogleIcon />
            Continue with Google
          </NexusButton>

          <div className="flex items-center gap-3">
            <Separator className="flex-1" />
            <span className="text-muted-foreground text-xs">
              or sign in with email
            </span>
            <Separator className="flex-1" />
          </div>

          {/* Form */}
          <form className="space-y-4" onSubmit={handleSubmit}>
            <FormGroup label="Username" required error={errors.username}>
              <NexusInput
                placeholder="your.username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
                className="h-11 rounded-xl"
              />
            </FormGroup>
            <FormGroup label="Password" required error={errors.password}>
              <NexusInput
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
                className="h-11 rounded-xl"
              />
            </FormGroup>
            <div className="flex items-center justify-between text-sm">
              <label className="text-muted-foreground flex cursor-pointer items-center gap-2">
                <input type="checkbox" className="border-border rounded" />
                Remember me
              </label>
              <Link
                to="/forgot-password"
                className="text-primary font-medium hover:underline"
              >
                Forgot?
              </Link>
            </div>
            <NexusButton
              className="h-11 w-full gap-2 rounded-xl"
              loading={loading}
              disabled={googleLoading}
            >
              <Sparkles className="h-4 w-4" />
              Sign In
            </NexusButton>
          </form>

          <p className="text-muted-foreground text-center text-sm">
            New here?{" "}
            <Link
              to="/register"
              className="text-primary font-medium hover:underline"
            >
              Create an account
            </Link>
          </p>
        </div>
      </motion.div>
    </div>
  );
}
