import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import {  NexusButton  } from "@casbin/ui";
import {  NexusInput  } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { Hexagon, CheckCircle, Lock, Eye, EyeOff } from "lucide-react";
import { toast } from "@casbin/ui";
import { motion, AnimatePresence } from "framer-motion";

const resetSchema = z.object({
  password: z.string().min(8, "Minimal 8 karakter"),
  confirmPassword: z.string().min(1, "Konfirmasi password wajib diisi"),
}).refine((d) => d.password === d.confirmPassword, { message: "Password tidak cocok", path: ["confirmPassword"] });

const requirements = [
  { label: "8+ chars", test: (p: string) => p.length >= 8 },
  { label: "Number", test: (p: string) => /\d/.test(p) },
  { label: "Uppercase", test: (p: string) => /[A-Z]/.test(p) },
  { label: "Special", test: (p: string) => /[!@#$%^&*]/.test(p) },
];

export default function ResetPasswordV3() {
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
      result.error.errors.forEach((err) => { if (err.path[0]) fe[String(err.path[0])] = err.message; });
      setErrors(fe); return;
    }
    setLoading(true);
    await new Promise((r) => setTimeout(r, 1200));
    toast.success("Password berhasil direset!");
    setDone(true);
    setLoading(false);
  };

  const strength = requirements.filter((r) => r.test(password)).length;

  return (
    <div className="min-h-screen flex items-center justify-center bg-foreground relative overflow-hidden">
      {/* Grid */}
      <div className="absolute inset-0 opacity-[0.04]" style={{
        backgroundImage: "linear-gradient(hsl(var(--background)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--background)) 1px, transparent 1px)",
        backgroundSize: "48px 48px",
      }} />
      <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 h-[400px] w-[400px] rounded-full bg-primary/10 blur-[120px]" />

      <motion.div initial={{ opacity: 0, y: 30 }} animate={{ opacity: 1, y: 0 }} className="relative z-10 w-full max-w-md mx-4">
        <div className="bg-background/5 backdrop-blur-md border border-background/10 rounded-2xl p-8 sm:p-10">
          <AnimatePresence mode="wait">
            {!done ? (
              <motion.div key="form" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="space-y-6">
                <div className="space-y-3">
                  <div className="flex items-center gap-2 mb-2">
                    <Hexagon className="h-7 w-7 text-primary" />
                    <span className="text-sm font-semibold text-background/60 tracking-widest uppercase">Nexus</span>
                  </div>
                  <h1 className="text-2xl font-bold text-background">Set new password</h1>
                  <p className="text-sm text-background/50">Create a strong, unique password.</p>
                </div>
                <form className="space-y-5" onSubmit={handleSubmit}>
                  <FormGroup label="New Password" required error={errors.password}>
                    <div className="relative">
                      <NexusInput type={showPassword ? "text" : "password"} placeholder="Enter new password" value={password}
                        onChange={(e) => setPassword(e.target.value)} disabled={loading}
                        className="h-11 bg-background/5 border-background/10 text-background placeholder:text-background/30 pr-10" />
                      <button type="button" onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-background/40 hover:text-background/70">
                        {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                  </FormGroup>
                  {password.length > 0 && (
                    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-2">
                      <div className="flex gap-1">
                        {[1, 2, 3, 4].map((i) => (
                          <div key={i} className={`h-1 flex-1 rounded-full transition-colors ${i <= strength ? "bg-primary" : "bg-background/10"}`} />
                        ))}
                      </div>
                      <div className="flex flex-wrap gap-1.5">
                        {requirements.map((req) => (
                          <span key={req.label} className={`text-xs px-2 py-0.5 rounded-full border transition-colors ${req.test(password) ? "border-primary/30 bg-primary/10 text-primary" : "border-background/10 text-background/30"}`}>
                            {req.label}
                          </span>
                        ))}
                      </div>
                    </motion.div>
                  )}
                  <FormGroup label="Confirm Password" required error={errors.confirmPassword}>
                    <div className="relative">
                      <NexusInput type={showConfirm ? "text" : "password"} placeholder="Confirm password" value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)} disabled={loading}
                        className="h-11 bg-background/5 border-background/10 text-background placeholder:text-background/30 pr-10" />
                      <button type="button" onClick={() => setShowConfirm(!showConfirm)}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-background/40 hover:text-background/70">
                        {showConfirm ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                    {confirmPassword.length > 0 && password === confirmPassword && (
                      <p className="text-xs text-primary mt-1 flex items-center gap-1"><CheckCircle className="h-3 w-3" /> Match</p>
                    )}
                  </FormGroup>
                  <NexusButton className="w-full h-11 gap-2 bg-primary hover:bg-primary/90" loading={loading}>
                    <Lock className="h-4 w-4" /> Reset Password
                  </NexusButton>
                </form>
                <Link to="/login" className="block text-center text-sm text-background/40 hover:text-background/60">Back to Sign In</Link>
              </motion.div>
            ) : (
              <motion.div key="success" initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} className="text-center space-y-6">
                <motion.div initial={{ scale: 0 }} animate={{ scale: 1 }} transition={{ type: "spring", delay: 0.2 }} className="flex justify-center">
                  <div className="h-20 w-20 rounded-full bg-primary/10 border-4 border-primary/20 flex items-center justify-center">
                    <CheckCircle className="h-10 w-10 text-primary" />
                  </div>
                </motion.div>
                <h2 className="text-2xl font-bold text-background">All done!</h2>
                <p className="text-sm text-background/50">Your password has been successfully reset.</p>
                <NexusButton className="w-full h-11 bg-primary hover:bg-primary/90" onClick={() => navigate("/login")}>Continue to Sign In</NexusButton>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </motion.div>
    </div>
  );
}
