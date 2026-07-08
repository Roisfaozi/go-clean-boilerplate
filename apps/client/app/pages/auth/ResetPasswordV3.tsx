import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { Hexagon, CheckCircle, Lock, Eye, EyeOff } from "lucide-react";
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
		<div className="bg-foreground relative flex min-h-screen items-center justify-center overflow-hidden">
			{/* Grid */}
			<div
				className="absolute inset-0 opacity-[0.04]"
				style={{
					backgroundImage:
						"linear-gradient(hsl(var(--background)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--background)) 1px, transparent 1px)",
					backgroundSize: "48px 48px",
				}}
			/>
			<div className="bg-primary/10 absolute top-1/3 left-1/2 h-[400px] w-[400px] -translate-x-1/2 -translate-y-1/2 rounded-full blur-[120px]" />

			<motion.div
				initial={{ opacity: 0, y: 30 }}
				animate={{ opacity: 1, y: 0 }}
				className="relative z-10 mx-4 w-full max-w-md"
			>
				<div className="bg-background/5 border-background/10 rounded-2xl border p-8 backdrop-blur-md sm:p-10">
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
									<div className="mb-2 flex items-center gap-2">
										<Hexagon className="text-primary h-7 w-7" />
										<span className="text-background/60 text-sm font-semibold tracking-widest uppercase">
											Nexus
										</span>
									</div>
									<h1 className="text-background text-2xl font-bold">
										Set new password
									</h1>
									<p className="text-background/50 text-sm">
										Create a strong, unique password.
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
												className="bg-background/5 border-background/10 text-background placeholder:text-background/30 h-11 pr-10"
											/>
											<button
												type="button"
												onClick={() => setShowPassword(!showPassword)}
												className="text-background/40 hover:text-background/70 absolute top-1/2 right-3 -translate-y-1/2"
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
														className={`h-1 flex-1 rounded-full transition-colors ${i <= strength ? "bg-primary" : "bg-background/10"}`}
													/>
												))}
											</div>
											<div className="flex flex-wrap gap-1.5">
												{requirements.map((req) => (
													<span
														key={req.label}
														className={`rounded-full border px-2 py-0.5 text-xs transition-colors ${req.test(password) ? "border-primary/30 bg-primary/10 text-primary" : "border-background/10 text-background/30"}`}
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
												className="bg-background/5 border-background/10 text-background placeholder:text-background/30 h-11 pr-10"
											/>
											<button
												type="button"
												onClick={() => setShowConfirm(!showConfirm)}
												className="text-background/40 hover:text-background/70 absolute top-1/2 right-3 -translate-y-1/2"
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
													<CheckCircle className="h-3 w-3" /> Match
												</p>
											)}
									</FormGroup>
									<NexusButton
										className="bg-primary hover:bg-primary/90 h-11 w-full gap-2"
										loading={loading}
									>
										<Lock className="h-4 w-4" /> Reset Password
									</NexusButton>
								</form>
								<Link
									to="/login"
									className="text-background/40 hover:text-background/60 block text-center text-sm"
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
									<div className="bg-primary/10 border-primary/20 flex h-20 w-20 items-center justify-center rounded-full border-4">
										<CheckCircle className="text-primary h-10 w-10" />
									</div>
								</motion.div>
								<h2 className="text-background text-2xl font-bold">
									All done!
								</h2>
								<p className="text-background/50 text-sm">
									Your password has been successfully reset.
								</p>
								<NexusButton
									className="bg-primary hover:bg-primary/90 h-11 w-full"
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
