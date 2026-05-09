import { useState } from "react";
import { Link } from "react-router";
import { z } from "zod";
import {  NexusButton  } from "@casbin/ui";
import {  NexusInput  } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { Hexagon, ArrowLeft, Mail, Send } from "lucide-react";
import { authApi } from "@/lib/api/auth";
import { toast } from "@casbin/ui";
import { motion, AnimatePresence } from "framer-motion";

const emailSchema = z.object({
  email: z.string().trim().email("Email tidak valid"),
});

export default function ForgotPasswordPage() {
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
    try {
      await authApi.forgotPassword(result.data.email);
    } catch {
      // Always show success to prevent email enumeration
    }
    toast.success("Link reset password telah dikirim ke email kamu");
    setSent(true);
    setLoading(false);
  };

  return (
    <div className="min-h-screen relative flex items-center justify-center bg-background overflow-hidden">
      {/* Decorative background circles */}
      <div className="absolute top-[-20%] right-[-10%] h-[600px] w-[600px] rounded-full bg-primary/5 blur-3xl" />
      <div className="absolute bottom-[-15%] left-[-10%] h-[500px] w-[500px] rounded-full bg-accent/5 blur-3xl" />
      
      {/* Grid lines */}
      <div className="absolute inset-0 opacity-[0.03]" style={{
        backgroundImage: "linear-gradient(hsl(var(--foreground)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--foreground)) 1px, transparent 1px)",
        backgroundSize: "64px 64px",
      }} />

      <motion.div
        initial={{ opacity: 0, y: 30, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
        className="relative z-10 w-full max-w-md mx-4"
      >
        {/* Back link */}
        <Link to="/login" className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors mb-8">
          <ArrowLeft className="h-4 w-4" /> Back to Sign In
        </Link>

        <div className="bg-card border border-border rounded-2xl shadow-xl p-8 sm:p-10 space-y-6 backdrop-blur-sm">
          <AnimatePresence mode="wait">
            {!sent ? (
              <motion.div
                key="form"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0, x: -20 }}
                className="space-y-6"
              >
                <div className="space-y-3">
                  <div className="h-12 w-12 rounded-xl bg-primary/10 flex items-center justify-center">
                    <Hexagon className="h-6 w-6 text-primary" />
                  </div>
                  <h1 className="text-2xl font-bold text-foreground">Forgot your password?</h1>
                  <p className="text-sm text-muted-foreground leading-relaxed">
                    No worries! Enter the email address associated with your account and we'll send you a link to reset your password.
                  </p>
                </div>

                <form className="space-y-4" onSubmit={handleSubmit}>
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
                  <NexusButton className="w-full h-11 gap-2" loading={loading}>
                    <Send className="h-4 w-4" /> Send Reset Link
                  </NexusButton>
                </form>
              </motion.div>
            ) : (
              <motion.div
                key="success"
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                className="space-y-6 text-center"
              >
                <motion.div
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  transition={{ type: "spring", stiffness: 200, damping: 15, delay: 0.2 }}
                  className="flex justify-center"
                >
                  <div className="h-20 w-20 rounded-full bg-success/10 border-4 border-success/20 flex items-center justify-center">
                    <Mail className="h-9 w-9 text-success" />
                  </div>
                </motion.div>
                <div className="space-y-2">
                  <h2 className="text-xl font-bold text-foreground">Check your email</h2>
                  <p className="text-sm text-muted-foreground">
                    We've sent a password reset link to
                  </p>
                  <p className="text-sm font-semibold text-foreground bg-muted rounded-lg py-2 px-4 inline-block">
                    {email}
                  </p>
                </div>
                <div className="space-y-3">
                  <NexusButton variant="outline" className="w-full h-11" onClick={() => { setSent(false); setLoading(false); }}>
                    Didn't receive? Send again
                  </NexusButton>
                  <p className="text-xs text-muted-foreground">
                    Check your spam folder if you don't see the email.
                  </p>
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  );
}
