import { useState } from "react";
import { Link, useNavigate } from "react-router";
import { z } from "zod";
import { NexusButton } from "@casbin/ui";
import { NexusInput } from "@casbin/ui";
import { FormGroup } from "@/components/patterns/form-group";
import { Hexagon, Check, ArrowRight } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { authApi } from "@/lib/api/auth";
import { toast } from "@casbin/ui";
import { Separator } from "@casbin/ui";
import { motion } from "framer-motion";

const registerSchema = z
	.object({
		name: z.string().trim().min(1, "Nama wajib diisi").max(100),
		email: z.string().trim().email("Email tidak valid").max(255),
		username: z.string().trim().min(3, "Minimal 3 karakter").max(50),
		password: z.string().min(8, "Minimal 8 karakter"),
		confirmPassword: z.string().min(1, "Konfirmasi password wajib diisi"),
	})
	.refine((d) => d.password === d.confirmPassword, {
		message: "Password tidak cocok",
		path: ["confirmPassword"],
	});

const GoogleIcon = () => (
	<svg className="h-4 w-4" viewBox="0 0 24 24">
		<path
			d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"
			fill="#4285F4"
		/>
		<path
			d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
			fill="#34A853"
		/>
		<path
			d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18A10.96 10.96 0 0 0 1 12c0 1.77.42 3.45 1.18 4.93l3.66-2.84z"
			fill="#FBBC05"
		/>
		<path
			d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
			fill="#EA4335"
		/>
	</svg>
);

const features = [
	"Unlimited projects & workspaces",
	"Role-based access control",
	"Real-time collaboration",
	"Advanced audit logging",
	"API management & endpoints",
];

export default function RegisterPage() {
	const [loading, setLoading] = useState(false);
	const [form, setForm] = useState({
		name: "",
		email: "",
		username: "",
		password: "",
		confirmPassword: "",
	});
	const [errors, setErrors] = useState<Record<string, string>>({});
	const navigate = useNavigate();
	const login = useAuthStore((s) => s.login);

	const update = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
		setForm((f) => ({ ...f, [field]: e.target.value }));

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setErrors({});
		const result = registerSchema.safeParse(form);
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
			const { confirmPassword: _, ...data } = result.data;
			const res = await authApi.register(
				data as {
					name: string;
					email: string;
					username: string;
					password: string;
				},
			);
			login(res.user);
			toast.success("Registrasi berhasil!");
			navigate("/");
		} catch {
			login({
				id: "new1",
				name: result.data.name,
				email: result.data.email,
				username: result.data.username,
			});
			navigate("/");
		} finally {
			setLoading(false);
		}
	};

	const handleGoogleSignup = () => {
		login({
			id: "g1",
			name: "Google User",
			email: "user@gmail.com",
			username: "googleuser",
		});
		navigate("/");
	};

	// Password strength
	const strength =
		form.password.length === 0
			? 0
			: form.password.length < 8
				? 1
				: form.password.length < 12
					? 2
					: 3;
	const strengthLabel = ["", "Weak", "Good", "Strong"][strength];
	const strengthColor = ["", "bg-danger", "bg-warning", "bg-success"][strength];

	return (
		<div className="flex min-h-screen">
			{/* Right side - gradient hero (reversed from login) */}
			<div className="from-secondary via-secondary to-accent relative hidden overflow-hidden bg-gradient-to-bl lg:flex lg:w-[45%]">
				{/* Pattern overlay */}
				<div
					className="absolute inset-0 opacity-[0.07]"
					style={{
						backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
					}}
				/>

				<div className="text-secondary-foreground relative z-10 flex w-full flex-col justify-center p-12">
					<motion.div
						initial={{ opacity: 0, y: 30 }}
						animate={{ opacity: 1, y: 0 }}
						transition={{ duration: 0.7 }}
						className="max-w-sm space-y-8"
					>
						<div>
							<h2 className="mb-3 text-3xl leading-tight font-bold">
								Start building
								<br />
								<span className="opacity-80">something amazing.</span>
							</h2>
							<p className="text-sm opacity-70">
								Join thousands of teams using NexusOS to manage their
								infrastructure.
							</p>
						</div>

						<div className="space-y-3">
							{features.map((feature, i) => (
								<motion.div
									key={feature}
									initial={{ opacity: 0, x: -20 }}
									animate={{ opacity: 1, x: 0 }}
									transition={{ delay: 0.3 + i * 0.1 }}
									className="flex items-center gap-3 text-sm"
								>
									<div className="bg-secondary-foreground/15 flex h-6 w-6 shrink-0 items-center justify-center rounded-full">
										<Check className="h-3.5 w-3.5" />
									</div>
									<span className="opacity-90">{feature}</span>
								</motion.div>
							))}
						</div>

						<div className="border-secondary-foreground/15 border-t pt-4">
							<p className="text-xs opacity-60">
								"NexusOS has transformed how we manage our team's permissions
								and projects."
							</p>
							<p className="mt-2 text-xs opacity-40">
								— Engineering Lead, TechCorp
							</p>
						</div>
					</motion.div>
				</div>
			</div>

			{/* Left side - form */}
			<div className="bg-background flex flex-1 items-center justify-center p-6 sm:p-12">
				<motion.div
					initial={{ opacity: 0, x: -20 }}
					animate={{ opacity: 1, x: 0 }}
					transition={{ duration: 0.5 }}
					className="w-full max-w-[440px] space-y-6"
				>
					<div className="mb-2 flex items-center gap-3">
						<Hexagon className="text-primary h-8 w-8" />
						<span className="text-foreground text-xl font-bold lg:hidden">
							NexusOS
						</span>
					</div>

					<div className="space-y-1">
						<h1 className="text-foreground text-3xl font-bold tracking-tight">
							Create account
						</h1>
						<p className="text-muted-foreground">
							Free forever. No credit card required.
						</p>
					</div>

					<NexusButton
						variant="outline"
						className="h-11 w-full gap-2"
						onClick={handleGoogleSignup}
						disabled={loading}
					>
						<GoogleIcon />
						Sign up with Google
					</NexusButton>

					<div className="flex items-center gap-3">
						<Separator className="flex-1" />
						<span className="text-muted-foreground text-xs tracking-wider uppercase">
							or
						</span>
						<Separator className="flex-1" />
					</div>

					<form className="space-y-4" onSubmit={handleSubmit}>
						<div className="grid grid-cols-2 gap-4">
							<FormGroup label="Full Name" required error={errors.name}>
								<NexusInput
									placeholder="John Doe"
									value={form.name}
									onChange={update("name")}
									disabled={loading}
									className="h-11"
								/>
							</FormGroup>
							<FormGroup label="Username" required error={errors.username}>
								<NexusInput
									placeholder="johndoe"
									value={form.username}
									onChange={update("username")}
									disabled={loading}
									className="h-11"
								/>
							</FormGroup>
						</div>
						<FormGroup label="Email" required error={errors.email}>
							<NexusInput
								type="email"
								placeholder="john@example.com"
								value={form.email}
								onChange={update("email")}
								disabled={loading}
								className="h-11"
							/>
						</FormGroup>
						<FormGroup label="Password" required error={errors.password}>
							<NexusInput
								type="password"
								placeholder="Min. 8 characters"
								value={form.password}
								onChange={update("password")}
								disabled={loading}
								className="h-11"
							/>
							{form.password.length > 0 && (
								<div className="mt-2 flex items-center gap-2">
									<div className="flex flex-1 gap-1">
										{[1, 2, 3].map((level) => (
											<div
												key={level}
												className={`h-1 flex-1 rounded-full transition-colors ${strength >= level ? strengthColor : "bg-muted"}`}
											/>
										))}
									</div>
									<span className="text-muted-foreground text-xs">
										{strengthLabel}
									</span>
								</div>
							)}
						</FormGroup>
						<FormGroup
							label="Confirm Password"
							required
							error={errors.confirmPassword}
						>
							<NexusInput
								type="password"
								placeholder="Confirm password"
								value={form.confirmPassword}
								onChange={update("confirmPassword")}
								disabled={loading}
								className="h-11"
							/>
						</FormGroup>

						<NexusButton className="h-11 w-full gap-2" loading={loading}>
							Create Account <ArrowRight className="h-4 w-4" />
						</NexusButton>
					</form>

					<p className="text-muted-foreground text-center text-sm">
						Already have an account?{" "}
						<Link
							to="/login"
							className="text-primary font-medium hover:underline"
						>
							Sign in
						</Link>
					</p>

					<p className="text-muted-foreground/60 text-center text-[11px]">
						By creating an account you agree to our Terms of Service and Privacy
						Policy
					</p>
				</motion.div>
			</div>
		</div>
	);
}
