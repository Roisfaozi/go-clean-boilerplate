import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import {
	Hexagon,
	CheckCircle,
	ShieldCheck,
	Lock,
	Eye,
	EyeOff,
} from "lucide-react";
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
	{ label: "At least 8 characters", test: (p: string) => p.length >= 8 },
	{ label: "Contains a number", test: (p: string) => /\d/.test(p) },
	{ label: "Contains uppercase", test: (p: string) => /[A-Z]/.test(p) },
	{ label: "Special character", test: (p: string) => /[!@#$%^&*]/.test(p) },
];

export default function ResetPasswordV1() {
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
							animate={{ rotate: [0, 360] }}
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
						className="bg-primary-foreground/10 border-primary-foreground/20 mb-6 flex h-20 w-20 items-center justify-center rounded-2xl border"
					>
						<ShieldCheck className="h-10 w-10" />
					</motion.div>
					<h2 className="mb-2 text-2xl font-bold">Secure Reset</h2>
					<p className="max-w-xs text-center text-sm opacity-50">
						Choose a strong password to keep your account safe.
					</p>
				</div>
			</div>

			{/* Right form panel */}
			<div className="bg-background flex flex-1 items-center justify-center p-6 sm:p-12">
				<AnimatePresence mode="wait">
					{!done ? (
						<motion.div
							key="form"
							initial={{ opacity: 0, y: 20 }}
							animate={{ opacity: 1, y: 0 }}
							exit={{ opacity: 0, y: -20 }}
							className="w-full max-w-[420px] space-y-8"
						>
							<div className="flex items-center gap-3 md:hidden">
								<Hexagon className="text-primary h-8 w-8" />
							</div>
							<div className="space-y-2">
								<h1 className="text-foreground text-3xl font-bold tracking-tight">
									Set new password
								</h1>
								<p className="text-muted-foreground">
									Your new password must be different from previously used
									passwords.
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
											className="h-11 pr-10"
										/>
										<button
											type="button"
											onClick={() => setShowPassword(!showPassword)}
											className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2 transition-colors"
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
										initial={{ opacity: 0, height: 0 }}
										animate={{ opacity: 1, height: "auto" }}
										className="grid grid-cols-2 gap-2"
									>
										{requirements.map((req) => {
											const met = req.test(password);
											return (
												<div
													key={req.label}
													className={`flex items-center gap-2 text-xs transition-colors ${met ? "text-primary" : "text-muted-foreground"}`}
												>
													<div
														className={`flex h-4 w-4 shrink-0 items-center justify-center rounded-full border ${met ? "border-primary bg-primary/10" : "border-border"}`}
													>
														{met && <CheckCircle className="h-3 w-3" />}
													</div>
													<span>{req.label}</span>
												</div>
											);
										})}
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
											placeholder="Confirm new password"
											value={confirmPassword}
											onChange={(e) => setConfirmPassword(e.target.value)}
											disabled={loading}
											className="h-11 pr-10"
										/>
										<button
											type="button"
											onClick={() => setShowConfirm(!showConfirm)}
											className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2 transition-colors"
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
								<NexusButton className="h-11 w-full gap-2" loading={loading}>
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
									<CheckCircle className="text-primary h-12 w-12" />
								</div>
							</motion.div>
							<h1 className="text-foreground text-3xl font-bold">All done!</h1>
							<p className="text-muted-foreground">
								Your password has been successfully reset.
							</p>
							<NexusButton
								className="mx-auto h-11 w-full max-w-xs"
								onClick={() => navigate("/login")}
							>
								Continue to Sign In
							</NexusButton>
						</motion.div>
					)}
				</AnimatePresence>
			</div>
		</div>
	);
}
