import { Link } from "react-router";
import {
	ArrowRight,
	Sparkles,
	ShieldCheck,
	KeySquare,
	Activity,
	Users,
} from "lucide-react";

export function LandingHero() {
	return (
		<section className="border-border relative overflow-hidden border-b">
			<div className="pointer-events-none absolute inset-0 overflow-hidden">
				<div className="bg-primary/15 absolute -top-24 left-1/4 h-[28rem] w-[28rem] rounded-full blur-3xl" />
				<div className="bg-accent/15 absolute -top-12 right-1/4 h-[28rem] w-[28rem] rounded-full blur-3xl" />
			</div>

			<div className="px-layout relative mx-auto grid max-w-7xl grid-cols-1 items-center gap-12 py-20 md:py-28 lg:grid-cols-12 lg:gap-10">
				<div className="lg:col-span-6">
					<div className="border-border bg-surface text-caption text-muted-foreground mb-5 inline-flex items-center gap-2 rounded-full border px-3 py-1.5 font-medium">
						<Sparkles className="text-primary h-3.5 w-3.5" />
						<span>Enterprise-ready workspace orchestration</span>
					</div>

					<h1 className="text-foreground text-4xl font-bold tracking-tight md:text-5xl lg:text-[56px] lg:leading-[1.05]">
						One platform to manage{" "}
						<span className="from-primary to-accent bg-gradient-to-r bg-clip-text text-transparent">
							operations, access, projects
						</span>{" "}
						and growth
					</h1>

					<p className="text-body-lg text-muted-foreground mt-5 max-w-xl">
						Bangun operasi bisnis, kontrol akses, manajemen project, integrasi
						API, dan observability dalam satu workspace yang konsisten dan
						scalable.
					</p>

					<div className="mt-8 flex flex-col gap-3 sm:flex-row">
						<Link
							to="/register"
							className="group h-btn bg-primary text-body text-primary-foreground hover:bg-primary-hover inline-flex items-center justify-center gap-2 rounded-md px-6 font-medium shadow-md transition-all hover:shadow-lg"
						>
							Start Free Trial
							<ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
						</Link>
						<a
							href="#pricing"
							className="h-btn border-border bg-background text-body text-foreground hover:bg-surface-hover inline-flex items-center justify-center rounded-md border px-6 font-medium transition-colors"
						>
							Book a Demo
						</a>
					</div>

					<ul className="text-caption text-muted-foreground mt-8 grid grid-cols-2 gap-3 sm:grid-cols-4">
						{[
							{ icon: ShieldCheck, label: "No credit card" },
							{ icon: KeySquare, label: "SSO ready" },
							{ icon: Users, label: "Role-aware" },
							{ icon: Activity, label: "Realtime" },
						].map(({ icon: Icon, label }) => (
							<li key={label} className="inline-flex items-center gap-1.5">
								<Icon className="text-secondary h-3.5 w-3.5" />
								{label}
							</li>
						))}
					</ul>
				</div>

				{/* Product mockup */}
				<div className="lg:col-span-6">
					<div className="border-border bg-card relative overflow-hidden rounded-xl border shadow-xl">
						<div className="border-border bg-surface flex items-center gap-1.5 border-b px-4 py-3">
							<span className="bg-danger/70 h-2.5 w-2.5 rounded-full" />
							<span className="bg-warning/70 h-2.5 w-2.5 rounded-full" />
							<span className="bg-success/70 h-2.5 w-2.5 rounded-full" />
							<span className="text-caption text-muted-foreground ml-3">
								app.nexusos.io / dashboard
							</span>
						</div>
						<div className="grid grid-cols-12 gap-3 p-4">
							{/* Side nav */}
							<div className="bg-surface col-span-3 space-y-1.5 rounded-md p-2">
								{[
									"Overview",
									"Users",
									"Projects",
									"Access",
									"Audit",
									"Settings",
								].map((l, i) => (
									<div
										key={l}
										className={`text-caption rounded px-2 py-1.5 ${
											i === 0
												? "bg-primary/10 text-primary font-medium"
												: "text-muted-foreground"
										}`}
									>
										{l}
									</div>
								))}
							</div>
							{/* Main */}
							<div className="col-span-9 space-y-3">
								<div className="grid grid-cols-3 gap-3">
									{[
										{
											label: "Active users",
											value: "2,847",
											track: "bg-primary/15",
											bar: "bg-primary",
										},
										{
											label: "API calls",
											value: "1.2M",
											track: "bg-secondary/15",
											bar: "bg-secondary",
										},
										{
											label: "Uptime",
											value: "99.98%",
											track: "bg-success/15",
											bar: "bg-success",
										},
									].map((s) => (
										<div
											key={s.label}
											className="border-border bg-background rounded-md border p-3"
										>
											<div className="text-caption text-muted-foreground">
												{s.label}
											</div>
											<div className="text-h3 text-foreground mt-1 font-semibold">
												{s.value}
											</div>
											<div
												className={`mt-2 h-1 w-full rounded-full ${s.track}`}
											>
												<div className={`h-1 w-2/3 rounded-full ${s.bar}`} />
											</div>
										</div>
									))}
								</div>
								<div className="border-border bg-background rounded-md border p-3">
									<div className="mb-2 flex items-center justify-between">
										<span className="text-caption text-foreground font-medium">
											Activity
										</span>
										<span className="text-caption text-muted-foreground">
											Last 24h
										</span>
									</div>
									<div className="flex h-20 items-end gap-1.5">
										{[40, 65, 30, 80, 55, 90, 70, 60, 85, 45, 75, 95].map(
											(h, i) => (
												<div
													key={i}
													className="from-primary/30 to-primary flex-1 rounded-t bg-gradient-to-t"
													style={{ height: `${h}%` }}
												/>
											),
										)}
									</div>
								</div>
								<div className="space-y-1.5">
									{[
										{ who: "Sarah", what: "granted admin role", when: "2m" },
										{ who: "Marcus", what: "deployed v2.4.0", when: "8m" },
										{ who: "Aisha", what: "invited 3 members", when: "14m" },
									].map((a, i) => (
										<div
											key={i}
											className="border-border bg-background text-caption flex items-center justify-between rounded border px-3 py-2"
										>
											<span className="text-foreground">
												<span className="font-medium">{a.who}</span>{" "}
												<span className="text-muted-foreground">{a.what}</span>
											</span>
											<span className="text-muted-foreground">
												{a.when} ago
											</span>
										</div>
									))}
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</section>
	);
}
