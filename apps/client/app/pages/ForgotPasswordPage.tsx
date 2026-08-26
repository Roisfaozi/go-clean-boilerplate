import { useState } from "react";
import { Link } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
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
		<div className="bg-background relative flex min-h-screen items-center justify-center overflow-hidden">
			{/* Decorative background circles */}
			<div className="bg-primary/5 absolute top-[-20%] right-[-10%] h-[600px] w-[600px] rounded-full blur-3xl" />
			<div className="bg-accent/5 absolute bottom-[-15%] left-[-10%] h-[500px] w-[500px] rounded-full blur-3xl" />

			{/* Grid lines */}
			<div
				className="absolute inset-0 opacity-[0.03]"
				style={{
					backgroundImage:
						"linear-gradient(hsl(var(--foreground)) 1px, transparent 1px), linear-gradient(90deg, hsl(var(--foreground)) 1px, transparent 1px)",
					backgroundSize: "64px 64px",
				}}
			/>

			<motion.div
				initial={{ opacity: 0, y: 30, scale: 0.98 }}
				animate={{ opacity: 1, y: 0, scale: 1 }}
				transition={{ duration: 0.5, ease: "easeOut" }}
				className="relative z-10 mx-4 w-full max-w-md"
			>
				{/* Back link */}
				<Link
					to="/login"
					className="text-muted-foreground hover:text-foreground mb-8 inline-flex items-center gap-2 text-sm transition-colors"
				>
					<ArrowLeft className="h-4 w-4" /> Back to Sign In
				</Link>

				<div className="bg-card border-border space-y-6 rounded-2xl border p-8 shadow-xl backdrop-blur-sm sm:p-10">
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
									<div className="bg-primary/10 flex h-12 w-12 items-center justify-center rounded-xl">
										<Hexagon className="text-primary h-6 w-6" />
									</div>
									<h1 className="text-foreground text-2xl font-bold">
										Forgot your password?
									</h1>
									<p className="text-muted-foreground text-sm leading-relaxed">
										No worries! Enter the email address associated with your
										account and we'll send you a link to reset your password.
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
									<NexusButton className="h-11 w-full gap-2" loading={loading}>
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
									<div className="bg-success/10 border-success/20 flex h-20 w-20 items-center justify-center rounded-full border-4">
										<Mail className="text-success h-9 w-9" />
									</div>
								</motion.div>
								<div className="space-y-2">
									<h2 className="text-foreground text-xl font-bold">
										Check your email
									</h2>
									<p className="text-muted-foreground text-sm">
										We've sent a password reset link to
									</p>
									<p className="text-foreground bg-muted inline-block rounded-lg px-4 py-2 text-sm font-semibold">
										{email}
									</p>
								</div>
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
									<p className="text-muted-foreground text-xs">
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
