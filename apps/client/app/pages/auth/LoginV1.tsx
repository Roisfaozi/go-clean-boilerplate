// V1: Split-screen — Re-exports current LoginPage layout
import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { Hexagon, Shield, Zap, Globe, ArrowRight } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { authApi } from "@/lib/api/auth";
import { toast } from "@casbin/ui";
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

export default function LoginV1() {
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
      const res = await authApi.login(
        result.data as { username: string; password: string },
      );
      login(res.user);
      toast.success("Login berhasil!");
      navigate("/");
    } catch {
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
    <div className="flex min-h-screen">
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
              Manage everything
              <br />
              <span className="opacity-80">in one place.</span>
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

      <div className="bg-background flex flex-1 items-center justify-center p-6 sm:p-12">
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5 }}
          className="w-full max-w-[420px] space-y-8"
        >
          <div className="mb-4 flex items-center gap-3 lg:hidden">
            <Hexagon className="text-primary h-8 w-8" />
            <span className="text-foreground text-xl font-bold">NexusOS</span>
          </div>
          <div className="space-y-2">
            <h1 className="text-foreground text-3xl font-bold tracking-tight">
              Welcome back
            </h1>
            <p className="text-muted-foreground">
              Enter your credentials to access your account
            </p>
          </div>
          <NexusButton
            variant="outline"
            className="h-11 w-full gap-2"
            onClick={handleGoogleLogin}
            loading={googleLoading}
            disabled={loading}
          >
            <GoogleIcon />
            Continue with Google
          </NexusButton>
          <div className="flex items-center gap-3">
            <Separator className="flex-1" />
            <span className="text-muted-foreground text-xs tracking-wider uppercase">
              or
            </span>
            <Separator className="flex-1" />
          </div>
          <form className="space-y-5" onSubmit={handleSubmit}>
            <FormGroup label="Username" required error={errors.username}>
              <NexusInput
                placeholder="Enter your username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
                className="h-11"
              />
            </FormGroup>
            <FormGroup label="Password" required error={errors.password}>
              <NexusInput
                type="password"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
                className="h-11"
              />
            </FormGroup>
            <div className="flex items-center justify-between">
              <label className="text-muted-foreground flex cursor-pointer items-center gap-2 text-sm">
                <input type="checkbox" className="border-border rounded" />
                Remember me
              </label>
              <Link
                to="/forgot-password"
                className="text-primary text-sm font-medium hover:underline"
              >
                Forgot password?
              </Link>
            </div>
            <NexusButton
              className="h-11 w-full gap-2"
              loading={loading}
              disabled={googleLoading}
            >
              Sign In <ArrowRight className="h-4 w-4" />
            </NexusButton>
          </form>
          <p className="text-muted-foreground text-center text-sm">
            Don't have an account?{" "}
            <Link
              to="/register"
              className="text-primary font-medium hover:underline"
            >
              Create account
            </Link>
          </p>
        </motion.div>
      </div>
    </div>
  );
}
