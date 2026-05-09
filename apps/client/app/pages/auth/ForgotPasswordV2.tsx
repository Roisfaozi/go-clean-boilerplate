import { useState } from "react";
import { Link } from "react-router";
import { z } from "zod";
import {  NexusButton  } from "@casbin/ui";
import {  NexusInput  } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { ArrowLeft, Mail, Send } from "lucide-react";
import { toast } from "@casbin/ui";
import { motion, AnimatePresence } from "framer-motion";

const emailSchema = z.object({ email: z.string().trim().email("Email tidak valid") });

export default function ForgotPasswordV2() {
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    const result = emailSchema.safeParse({ email });
    if (!result.success) { setError(result.error.errors[0]?.message || "Email tidak valid"); return; }
    setLoading(true);
    await new Promise((r) => setTimeout(r, 1200));
    toast.success("Link reset password telah dikirim");
    setSent(true);
    setLoading(false);
  };

  return (
    <div className="min-h-screen flex items-center justify-center relative overflow-hidden bg-background">
      {/* Animated gradient background */}
      <div className="absolute inset-0">
        <motion.div animate={{ rotate: [0, 360] }} transition={{ duration: 30, repeat: Infinity, ease: "linear" }}
          className="absolute inset-[-50%]"
          style={{ background: "conic-gradient(from 0deg, hsl(var(--primary)/0.15), hsl(var(--accent)/0.15), hsl(var(--secondary)/0.15), hsl(var(--primary)/0.15))" }} />
      </div>
      <div className="absolute inset-0 bg-background/60 backdrop-blur-2xl" />

      <motion.div initial={{ opacity: 0, y: 30, scale: 0.96 }} animate={{ opacity: 1, y: 0, scale: 1 }} transition={{ duration: 0.5 }}
        className="relative z-10 w-full max-w-md mx-4">
        <div className="bg-card/70 backdrop-blur-xl border border-border/50 rounded-3xl shadow-2xl p-8 sm:p-10">
          <AnimatePresence mode="wait">
            {!sent ? (
              <motion.div key="form" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0, x: -20 }} className="space-y-6">
                <Link to="/login" className="inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors">
                  <ArrowLeft className="h-4 w-4" /> Back to Sign In
                </Link>
                <div className="space-y-2">
                  <div className="h-14 w-14 rounded-2xl bg-gradient-to-br from-primary/20 to-accent/20 border border-primary/20 flex items-center justify-center mb-4">
                    <Mail className="h-7 w-7 text-primary" />
                  </div>
                  <h1 className="text-2xl font-bold text-foreground">Forgot password?</h1>
                  <p className="text-sm text-muted-foreground leading-relaxed">
                    No worries! Enter your email and we'll send you a recovery link.
                  </p>
                </div>
                <form className="space-y-4" onSubmit={handleSubmit}>
                  <FormGroup label="Email" required error={error}>
                    <NexusInput type="email" placeholder="you@company.com" value={email} onChange={(e) => setEmail(e.target.value)} disabled={loading}
                      className="h-12 rounded-xl bg-background/50 border-border/50 backdrop-blur-sm" />
                  </FormGroup>
                  <NexusButton className="w-full h-12 rounded-xl gap-2 bg-gradient-to-r from-primary to-accent text-primary-foreground" loading={loading}>
                    <Send className="h-4 w-4" /> Send Reset Link
                  </NexusButton>
                </form>
              </motion.div>
            ) : (
              <motion.div key="success" initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="space-y-6 text-center">
                <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }} transition={{ type: "spring", stiffness: 200, damping: 15, delay: 0.2 }} className="flex justify-center">
                  <div className="h-20 w-20 rounded-full bg-gradient-to-br from-primary/20 to-accent/20 border-4 border-primary/10 flex items-center justify-center">
                    <Mail className="h-9 w-9 text-primary" />
                  </div>
                </motion.div>
                <h2 className="text-xl font-bold text-foreground">Check your inbox</h2>
                <p className="text-sm text-muted-foreground">We sent a reset link to <span className="font-semibold text-foreground">{email}</span></p>
                <div className="space-y-3">
                  <NexusButton variant="outline" className="w-full h-11 rounded-xl" onClick={() => { setSent(false); setLoading(false); }}>Didn't receive? Resend</NexusButton>
                  <p className="text-xs text-muted-foreground">Check spam folder if you don't see it.</p>
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  );
}
