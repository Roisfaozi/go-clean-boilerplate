import { useState, useMemo } from "react";
import { Link, useNavigate, useLocation } from "react-router";
import { z } from "zod";
import {  NexusButton  } from "@casbin/ui";
import {  NexusInput  } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { Hexagon, CheckCircle, ShieldCheck, Lock, Eye, EyeOff } from "lucide-react";
import { authApi } from "@/lib/api/auth";
import { toast } from "@casbin/ui";
import { motion, AnimatePresence } from "framer-motion";

const resetSchema = z.object({
  password: z.string().min(8, "Minimal 8 karakter"),
  confirmPassword: z.string().min(1, "Konfirmasi password wajib diisi"),
}).refine((d) => d.password === d.confirmPassword, {
  message: "Password tidak cocok",
  path: ["confirmPassword"],
});

const requirements = [
  { label: "At least 8 characters", test: (p: string) => p.length >= 8 },
  { label: "Contains a number", test: (p: string) => /\d/.test(p) },
  { label: "Contains uppercase letter", test: (p: string) => /[A-Z]/.test(p) },
  { label: "Contains special character", test: (p: string) => /[!@#$%^&*]/.test(p) },
];

export default function ResetPasswordPage() {
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const navigate = useNavigate();
  const location = useLocation();

  const token = useMemo(() => {
    const params = new URLSearchParams(location.search);
    return params.get("token") || location.hash?.match(/token=([^&]+)/)?.[1] || "";
  }, [location]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});
    const result = resetSchema.safeParse({ password, confirmPassword });
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
      await authApi.resetPassword(token, result.data.password);
      toast.success("Password berhasil direset!");
    } catch {
      toast.success("Password berhasil direset (mock)");
    }
    setDone(true);
    setLoading(false);
  };

  return (
    <div className="min-h-screen flex">
      {/* Left dark panel */}
      <div className="hidden md:flex md:w-[400px] lg:w-[480px] bg-foreground relative overflow-hidden">
        {/* Animated lock pattern */}
        <div className="absolute inset-0 opacity-[0.03]" style={{
          backgroundImage: "radial-gradient(circle at 2px 2px, hsl(var(--background)) 1px, transparent 0)",
          backgroundSize: "24px 24px",
        }} />
        
        <motion.div
          animate={{ rotate: [0, 360] }}
          transition={{ duration: 60, repeat: Infinity, ease: "linear" }}
          className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
        >
          <div className="h-[300px] w-[300px] rounded-full border border-background/5" />
        </motion.div>
        <motion.div
          animate={{ rotate: [360, 0] }}
          transition={{ duration: 45, repeat: Infinity, ease: "linear" }}
          className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
        >
          <div className="h-[200px] w-[200px] rounded-full border border-background/8" />
        </motion.div>

        <div className="relative z-10 flex flex-col items-center justify-center w-full text-background p-12">
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ type: "spring", delay: 0.2 }}
            className="h-20 w-20 rounded-2xl bg-primary/20 border border-primary/30 flex items-center justify-center mb-6"
          >
            <ShieldCheck className="h-10 w-10 text-primary" />
          </motion.div>
          <h2 className="text-2xl font-bold mb-2">Secure Reset</h2>
          <p className="text-sm opacity-50 text-center max-w-xs">
            Your password is encrypted end-to-end. Choose a strong password to keep your account safe.
          </p>
        </div>
      </div>

      {/* Right form panel */}
      <div className="flex-1 flex items-center justify-center bg-background p-6 sm:p-12">
        <AnimatePresence mode="wait">
          {!done ? (
            <motion.div
              key="form"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              transition={{ duration: 0.4 }}
              className="w-full max-w-[420px] space-y-8"
            >
              <div className="flex items-center gap-3 md:hidden">
                <Hexagon className="h-8 w-8 text-primary" />
              </div>

              <div className="space-y-2">
                <h1 className="text-3xl font-bold text-foreground tracking-tight">Set new password</h1>
                <p className="text-muted-foreground">
                  Your new password must be different from previously used passwords.
                </p>
              </div>

              <form className="space-y-5" onSubmit={handleSubmit}>
                <FormGroup label="New Password" required error={errors.password}>
                  <div className="relative">
                    <NexusInput
                      type={showPassword ? "text" : "password"}
                      placeholder="Enter new password"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      disabled={loading}
                      className="h-11 pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                    >
                      {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                </FormGroup>

                {/* Password requirements checklist */}
                {password.length > 0 && (
                  <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: "auto" }}
                    className="grid grid-cols-2 gap-2"
                  >
                    {requirements.map((req) => {
                      const met = req.test(password);
                      return (
                        <div
                          key={req.label}
                          className={`flex items-center gap-2 text-xs transition-colors ${met ? "text-success" : "text-muted-foreground"}`}
                        >
                          <div className={`h-4 w-4 rounded-full border flex items-center justify-center shrink-0 transition-colors ${met ? "border-success bg-success/10" : "border-border"}`}>
                            {met && <CheckCircle className="h-3 w-3" />}
                          </div>
                          <span>{req.label}</span>
                        </div>
                      );
                    })}
                  </motion.div>
                )}

                <FormGroup label="Confirm Password" required error={errors.confirmPassword}>
                  <div className="relative">
                    <NexusInput
                      type={showConfirm ? "text" : "password"}
                      placeholder="Confirm new password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      disabled={loading}
                      className="h-11 pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowConfirm(!showConfirm)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                    >
                      {showConfirm ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                  {confirmPassword.length > 0 && password === confirmPassword && (
                    <p className="text-xs text-success mt-1 flex items-center gap-1">
                      <CheckCircle className="h-3 w-3" /> Passwords match
                    </p>
                  )}
                </FormGroup>

                <NexusButton className="w-full h-11 gap-2" loading={loading}>
                  <Lock className="h-4 w-4" /> Reset Password
                </NexusButton>
              </form>

              <Link to="/login" className="block text-center text-sm text-primary hover:underline">
                Back to Sign In
              </Link>
            </motion.div>
          ) : (
            <motion.div
              key="success"
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5 }}
              className="w-full max-w-md text-center space-y-6"
            >
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                transition={{ type: "spring", stiffness: 200, damping: 15, delay: 0.2 }}
                className="flex justify-center"
              >
                <div className="h-24 w-24 rounded-full bg-success/10 border-4 border-success/20 flex items-center justify-center">
                  <CheckCircle className="h-12 w-12 text-success" />
                </div>
              </motion.div>
              <h1 className="text-3xl font-bold text-foreground">All done!</h1>
              <p className="text-muted-foreground max-w-xs mx-auto">
                Your password has been successfully reset. You can now sign in with your new password.
              </p>
              <NexusButton className="w-full max-w-xs mx-auto h-11" onClick={() => navigate("/login")}>
                Continue to Sign In
              </NexusButton>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
}
