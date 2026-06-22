import { useState } from "react";
import { Link } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { ArrowLeft, Mail, Send } from "lucide-react";
import { toast } from "@casbin/ui";
import { motion, AnimatePresence } from "framer-motion";

const emailSchema = z.object({
  email: z.string().trim().email("Email tidak valid"),
});

export default function ForgotPasswordV2() {
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
    toast.success("Link reset password telah dikirim");
    setSent(true);
    setLoading(false);
  };

  return (
    <div className="bg-background relative flex min-h-screen items-center justify-center overflow-hidden">
      {/* Animated gradient background */}
      <div className="absolute inset-0">
        <motion.div
          animate={{ rotate: [0, 360] }}
          transition={{ duration: 30, repeat: Infinity, ease: "linear" }}
          className="absolute inset-[-50%]"
          style={{
            background:
              "conic-gradient(from 0deg, hsl(var(--primary)/0.15), hsl(var(--accent)/0.15), hsl(var(--secondary)/0.15), hsl(var(--primary)/0.15))",
          }}
        />
      </div>
      <div className="bg-background/60 absolute inset-0 backdrop-blur-2xl" />

      <motion.div
        initial={{ opacity: 0, y: 30, scale: 0.96 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 0.5 }}
        className="relative z-10 mx-4 w-full max-w-md"
      >
        <div className="bg-card/70 border-border/50 rounded-3xl border p-8 shadow-2xl backdrop-blur-xl sm:p-10">
          <AnimatePresence mode="wait">
            {!sent ? (
              <motion.div
                key="form"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0, x: -20 }}
                className="space-y-6"
              >
                <Link
                  to="/login"
                  className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 text-sm transition-colors"
                >
                  <ArrowLeft className="h-4 w-4" /> Back to Sign In
                </Link>
                <div className="space-y-2">
                  <div className="from-primary/20 to-accent/20 border-primary/20 mb-4 flex h-14 w-14 items-center justify-center rounded-2xl border bg-gradient-to-br">
                    <Mail className="text-primary h-7 w-7" />
                  </div>
                  <h1 className="text-foreground text-2xl font-bold">
                    Forgot password?
                  </h1>
                  <p className="text-muted-foreground text-sm leading-relaxed">
                    No worries! Enter your email and we'll send you a recovery
                    link.
                  </p>
                </div>
                <form className="space-y-4" onSubmit={handleSubmit}>
                  <FormGroup label="Email" required error={error}>
                    <NexusInput
                      type="email"
                      placeholder="you@company.com"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      disabled={loading}
                      className="bg-background/50 border-border/50 h-12 rounded-xl backdrop-blur-sm"
                    />
                  </FormGroup>
                  <NexusButton
                    className="from-primary to-accent text-primary-foreground h-12 w-full gap-2 rounded-xl bg-gradient-to-r"
                    loading={loading}
                  >
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
                  transition={{
                    type: "spring",
                    stiffness: 200,
                    damping: 15,
                    delay: 0.2,
                  }}
                  className="flex justify-center"
                >
                  <div className="from-primary/20 to-accent/20 border-primary/10 flex h-20 w-20 items-center justify-center rounded-full border-4 bg-gradient-to-br">
                    <Mail className="text-primary h-9 w-9" />
                  </div>
                </motion.div>
                <h2 className="text-foreground text-xl font-bold">
                  Check your inbox
                </h2>
                <p className="text-muted-foreground text-sm">
                  We sent a reset link to{" "}
                  <span className="text-foreground font-semibold">{email}</span>
                </p>
                <div className="space-y-3">
                  <NexusButton
                    variant="outline"
                    className="h-11 w-full rounded-xl"
                    onClick={() => {
                      setSent(false);
                      setLoading(false);
                    }}
                  >
                    Didn't receive? Resend
                  </NexusButton>
                  <p className="text-muted-foreground text-xs">
                    Check spam folder if you don't see it.
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
