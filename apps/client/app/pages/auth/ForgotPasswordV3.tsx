import { useState } from "react";
import { Link } from "react-router";
import { z } from "zod";
import {  NexusButton  } from "@casbin/ui";
import {  NexusInput  } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { Hexagon, ArrowLeft, Mail, Send } from "lucide-react";
import { toast } from "@casbin/ui";
import { motion, AnimatePresence } from "framer-motion";

const emailSchema = z.object({ email: z.string().trim().email("Email tidak valid") });

export default function ForgotPasswordV3() {
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
    <div className="min-h-screen flex items-center justify-center bg-foreground relative overflow-hidden">
      {/* Grid pattern */}
      <div className="absolute inset-0 opacity-[0.04]" style={{
        backgroundImage: "linear-gradient(hsl(var(--background)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--background)) 1px, transparent 1px)",
        backgroundSize: "48px 48px",
      }} />
      {/* Glow */}
      <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 h-[400px] w-[400px] rounded-full bg-primary/10 blur-[120px]" />

      <motion.div initial={{ opacity: 0, y: 30 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }}
        className="relative z-10 w-full max-w-md mx-4">
        <div className="bg-background/5 backdrop-blur-md border border-background/10 rounded-2xl p-8 sm:p-10">
          <AnimatePresence mode="wait">
            {!sent ? (
              <motion.div key="form" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0, x: -20 }} className="space-y-6">
                <Link to="/login" className="inline-flex items-center gap-2 text-sm text-background/50 hover:text-background/80 transition-colors">
                  <ArrowLeft className="h-4 w-4" /> Back to Sign In
                </Link>
                <div className="space-y-3">
                  <div className="flex items-center gap-2 mb-2">
                    <Hexagon className="h-7 w-7 text-primary" />
                    <span className="text-sm font-semibold text-background/60 tracking-widest uppercase">Nexus</span>
                  </div>
                  <h1 className="text-2xl font-bold text-background">Forgot password?</h1>
                  <p className="text-sm text-background/50">Enter your email to receive a password reset link.</p>
                </div>
                <form className="space-y-4" onSubmit={handleSubmit}>
                  <FormGroup label="Email" required error={error}>
                    <NexusInput type="email" placeholder="you@company.com" value={email} onChange={(e) => setEmail(e.target.value)} disabled={loading}
                      className="h-11 bg-background/5 border-background/10 text-background placeholder:text-background/30 focus:border-primary" />
                  </FormGroup>
                  <NexusButton className="w-full h-11 gap-2 bg-primary hover:bg-primary/90" loading={loading}>
                    <Send className="h-4 w-4" /> Send Reset Link
                  </NexusButton>
                </form>
                <p className="text-xs text-background/30 text-center">
                  Protected by enterprise-grade encryption
                </p>
              </motion.div>
            ) : (
              <motion.div key="success" initial={{ opacity: 0, x: 20 }} animate={{ opacity: 1, x: 0 }} className="space-y-6 text-center">
                <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }} transition={{ type: "spring", stiffness: 200, damping: 15, delay: 0.2 }} className="flex justify-center">
                  <div className="h-20 w-20 rounded-full bg-primary/10 border-4 border-primary/20 flex items-center justify-center">
                    <Mail className="h-9 w-9 text-primary" />
                  </div>
                </motion.div>
                <h2 className="text-xl font-bold text-background">Check your email</h2>
                <p className="text-sm text-background/50">Reset link sent to <span className="font-semibold text-background/80">{email}</span></p>
                <div className="space-y-3">
                  <NexusButton variant="outline" className="w-full h-11 border-background/10 text-background/70 hover:bg-background/5"
                    onClick={() => { setSent(false); setLoading(false); }}>Didn't receive? Resend</NexusButton>
                  <p className="text-xs text-background/30">Check spam folder if you don't see it.</p>
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  );
}
