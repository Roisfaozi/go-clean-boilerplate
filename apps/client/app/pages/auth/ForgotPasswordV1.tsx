import { useState } from "react";
import { Link } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import {
  Hexagon,
  ArrowLeft,
  Mail,
  Send,
  KeyRound,
  Shield,
  Lock,
} from "lucide-react";
import { toast } from "@casbin/ui";
import { motion, AnimatePresence } from "framer-motion";

const emailSchema = z.object({
  email: z.string().trim().email("Email tidak valid"),
});

export default function ForgotPasswordV1() {
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    const result = emailSchema.safeParse({ email });
    if (!result.success) {
      setError(result.error.errors[0]?.message || "Email tidak valid");
      return;
    }
    setLoading(true);
    await new Promise((r) => setTimeout(r, 1200));
    toast.success("Link reset password telah dikirim ke email kamu");
    setSent(true);
    setLoading(false);
  };

  return (
    <div className="flex min-h-screen">
      {/* Left branding panel */}
      <div className="bg-primary relative hidden overflow-hidden md:flex md:w-[400px] lg:w-[480px]">
        <div className="absolute inset-0 opacity-10">
          {[...Array(5)].map((_, i) => (
            <motion.div
              key={i}
              className="border-primary-foreground/20 absolute rounded-full border"
              style={{
                width: 100 + i * 80,
                height: 100 + i * 80,
                top: "50%",
                left: "50%",
                x: "-50%",
                y: "-50%",
              }}
              animate={{ rotate: [0, 360], scale: [1, 1.1, 1] }}
              transition={{
                duration: 20 + i * 5,
                repeat: Infinity,
                ease: "linear",
              }}
            />
          ))}
        </div>
        <div className="text-primary-foreground relative z-10 flex w-full flex-col items-center justify-center p-12">
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ type: "spring", delay: 0.2 }}
            className="bg-primary-foreground/10 border-primary-foreground/20 mb-8 flex h-20 w-20 items-center justify-center rounded-2xl border"
          >
            <KeyRound className="h-10 w-10" />
          </motion.div>
          <h2 className="mb-3 text-2xl font-bold">Password Recovery</h2>
          <p className="mb-10 max-w-xs text-center text-sm opacity-70">
            Don't worry, it happens to the best of us. We'll help you get back
            into your account.
          </p>
          <div className="w-full max-w-xs space-y-4">
            {[
              { icon: Mail, text: "Enter your email address" },
              { icon: Shield, text: "Receive a secure reset link" },
              { icon: Lock, text: "Create a new password" },
            ].map((step, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.4 + i * 0.15 }}
                className="bg-primary-foreground/5 flex items-center gap-3 rounded-lg p-3"
              >
                <div className="bg-primary-foreground/10 flex h-8 w-8 shrink-0 items-center justify-center rounded-full">
                  <step.icon className="h-4 w-4" />
                </div>
                <span className="text-sm opacity-80">{step.text}</span>
              </motion.div>
            ))}
          </div>
        </div>
      </div>

      {/* Right form panel */}
      <div className="bg-background flex flex-1 items-center justify-center p-6 sm:p-12">
        <AnimatePresence mode="wait">
          {!sent ? (
            <motion.div
              key="form"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="w-full max-w-[420px] space-y-8"
            >
              <Link
                to="/login"
                className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 text-sm transition-colors"
              >
                <ArrowLeft className="h-4 w-4" /> Back to Sign In
              </Link>
              <div className="space-y-2">
                <div className="mb-4 flex items-center gap-3 md:hidden">
                  <Hexagon className="text-primary h-8 w-8" />
                </div>
                <h1 className="text-foreground text-3xl font-bold tracking-tight">
                  Forgot password?
                </h1>
                <p className="text-muted-foreground">
                  Enter the email associated with your account and we'll send a
                  reset link.
                </p>
              </div>
              <form className="space-y-5" onSubmit={handleSubmit}>
                <FormGroup label="Email address" required error={error}>
                  <NexusInput
                    type="email"
                    placeholder="you@company.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    disabled={loading}
                    className="h-11"
                  />
                </FormGroup>
                <NexusButton className="h-11 w-full gap-2" loading={loading}>
                  <Send className="h-4 w-4" /> Send Reset Link
                </NexusButton>
              </form>
            </motion.div>
          ) : (
            <motion.div
              key="success"
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              className="w-full max-w-md space-y-6 text-center"
            >
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{
                  type: "spring",
                  stiffness: 200,
                  damping: 15,
                  delay: 0.2,
                }}
                className="flex justify-center"
              >
                <div className="bg-primary/10 border-primary/20 flex h-24 w-24 items-center justify-center rounded-full border-4">
                  <Mail className="text-primary h-12 w-12" />
                </div>
              </motion.div>
              <h2 className="text-foreground text-2xl font-bold">
                Check your email
              </h2>
              <p className="text-muted-foreground">
                We've sent a password reset link to{" "}
                <span className="text-foreground font-semibold">{email}</span>
              </p>
              <div className="space-y-3">
                <NexusButton
                  variant="outline"
                  className="h-11 w-full"
                  onClick={() => {
                    setSent(false);
                    setLoading(false);
                  }}
                >
                  Didn't receive? Send again
                </NexusButton>
                <Link
                  to="/login"
                  className="text-primary block text-sm hover:underline"
                >
                  Back to Sign In
                </Link>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
}
