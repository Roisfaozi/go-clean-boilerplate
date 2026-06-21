import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { CheckCircle, Lock, Eye, EyeOff, ShieldCheck } from "lucide-react";
import { toast } from "@casbin/ui";
import { motion, AnimatePresence } from "framer-motion";

const resetSchema = z
  .object({
    password: z.string().min(8, "Minimal 8 karakter"),
    confirmPassword: z.string().min(1, "Konfirmasi password wajib diisi"),
  })
  .refine((d) => d.password === d.confirmPassword, {
    message: "Password tidak cocok",
    path: ["confirmPassword"],
  });

const requirements = [
  { label: "8+ characters", test: (p: string) => p.length >= 8 },
  { label: "Number", test: (p: string) => /\d/.test(p) },
  { label: "Uppercase", test: (p: string) => /[A-Z]/.test(p) },
  { label: "Special char", test: (p: string) => /[!@#$%^&*]/.test(p) },
];

export default function ResetPasswordV2() {
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});
    const result = resetSchema.safeParse({ password, confirmPassword });
    if (!result.success) {
      const fe: Record<string, string> = {};
      result.error.errors.forEach((err) => {
        if (err.path[0]) fe[String(err.path[0])] = err.message;
      });
      setErrors(fe);
      return;
    }
    setLoading(true);
    await new Promise((r) => setTimeout(r, 1200));
    toast.success("Password berhasil direset!");
    setDone(true);
    setLoading(false);
  };

  const strength = requirements.filter((r) => r.test(password)).length;

  return (
    <div className="bg-background relative flex min-h-screen items-center justify-center overflow-hidden">
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
            {!done ? (
              <motion.div
                key="form"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="space-y-6"
              >
                <div className="space-y-3">
                  <div className="from-primary/20 to-accent/20 border-primary/20 flex h-14 w-14 items-center justify-center rounded-2xl border bg-gradient-to-br">
                    <ShieldCheck className="text-primary h-7 w-7" />
                  </div>
                  <h1 className="text-foreground text-2xl font-bold">
                    Set new password
                  </h1>
                  <p className="text-muted-foreground text-sm">
                    Create a strong password for your account.
                  </p>
                </div>
                <form className="space-y-5" onSubmit={handleSubmit}>
                  <FormGroup
                    label="New Password"
                    required
                    error={errors.password}
                  >
                    <div className="relative">
                      <NexusInput
                        type={showPassword ? "text" : "password"}
                        placeholder="Enter new password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        disabled={loading}
                        className="bg-background/50 border-border/50 h-12 rounded-xl pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2"
                      >
                        {showPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                  </FormGroup>
                  {password.length > 0 && (
                    <motion.div
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      className="space-y-2"
                    >
                      <div className="flex gap-1">
                        {[1, 2, 3, 4].map((i) => (
                          <div
                            key={i}
                            className={`h-1.5 flex-1 rounded-full transition-colors ${i <= strength ? "bg-primary" : "bg-border"}`}
                          />
                        ))}
                      </div>
                      <div className="flex flex-wrap gap-2">
                        {requirements.map((req) => (
                          <span
                            key={req.label}
                            className={`rounded-full border px-2 py-0.5 text-xs transition-colors ${req.test(password) ? "border-primary/30 bg-primary/10 text-primary" : "border-border text-muted-foreground"}`}
                          >
                            {req.label}
                          </span>
                        ))}
                      </div>
                    </motion.div>
                  )}
                  <FormGroup
                    label="Confirm Password"
                    required
                    error={errors.confirmPassword}
                  >
                    <div className="relative">
                      <NexusInput
                        type={showConfirm ? "text" : "password"}
                        placeholder="Confirm password"
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        disabled={loading}
                        className="bg-background/50 border-border/50 h-12 rounded-xl pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowConfirm(!showConfirm)}
                        className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2"
                      >
                        {showConfirm ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </button>
                    </div>
                    {confirmPassword.length > 0 &&
                      password === confirmPassword && (
                        <p className="text-primary mt-1 flex items-center gap-1 text-xs">
                          <CheckCircle className="h-3 w-3" /> Passwords match
                        </p>
                      )}
                  </FormGroup>
                  <NexusButton
                    className="from-primary to-accent text-primary-foreground h-12 w-full gap-2 rounded-xl bg-gradient-to-r"
                    loading={loading}
                  >
                    <Lock className="h-4 w-4" /> Reset Password
                  </NexusButton>
                </form>
                <Link
                  to="/login"
                  className="text-primary block text-center text-sm hover:underline"
                >
                  Back to Sign In
                </Link>
              </motion.div>
            ) : (
              <motion.div
                key="success"
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="space-y-6 text-center"
              >
                <motion.div
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  transition={{ type: "spring", delay: 0.2 }}
                  className="flex justify-center"
                >
                  <div className="from-primary/20 to-accent/20 border-primary/10 flex h-20 w-20 items-center justify-center rounded-full border-4 bg-gradient-to-br">
                    <CheckCircle className="text-primary h-10 w-10" />
                  </div>
                </motion.div>
                <h2 className="text-foreground text-2xl font-bold">
                  All done!
                </h2>
                <p className="text-muted-foreground text-sm">
                  Your password has been successfully reset.
                </p>
                <NexusButton
                  className="h-12 w-full rounded-xl"
                  onClick={() => navigate("/login")}
                >
                  Continue to Sign In
                </NexusButton>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  );
}
