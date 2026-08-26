import { Link } from "react-router";
import {
	ArrowUpRight,
	Users,
	Shield,
	BarChart3,
	Upload,
	Building2,
	Activity,
} from "lucide-react";

const products = [
	{
		icon: Users,
		title: "User Management",
		description: "Provision users, manage roles, and audit activity at scale.",
		href: "/users",
		accent: "primary",
	},
	{
		icon: Shield,
		title: "Access Control",
		description: "RBAC with fine-grained permissions and inheritance trees.",
		href: "/permissions",
		accent: "accent",
	},
	{
		icon: BarChart3,
		title: "System Insights",
		description: "Real-time analytics, charts, and metrics across your stack.",
		href: "/system-insights",
		accent: "secondary",
	},
	{
		icon: Upload,
		title: "Resumable Uploads",
		description: "TUS-powered chunked uploads with progress tracking.",
		href: "/uploads",
		accent: "info",
	},
	{
		icon: Building2,
		title: "Multi-Workspace",
		description: "Organizations, projects, and member roles built-in.",
		href: "/workspace",
		accent: "warning",
	},
	{
		icon: Activity,
		title: "System Health",
		description: "Live monitoring with audit logs and uptime visibility.",
		href: "/system-health",
		accent: "success",
	},
];

const accentMap: Record<string, string> = {
	primary: "bg-primary/10 text-primary",
	accent: "bg-accent/10 text-accent",
	secondary: "bg-secondary/10 text-secondary",
	info: "bg-info/10 text-info",
	warning: "bg-warning/10 text-warning",
	success: "bg-success/10 text-success",
};

export function LandingProducts() {
	return (
		<section
			id="products"
			className="border-border bg-surface/50 border-b py-20 md:py-28"
		>
			<div className="mx-auto max-w-7xl px-6">
				<div className="mx-auto mb-14 max-w-2xl text-center">
					<span className="text-primary text-sm font-semibold tracking-wider uppercase">
						Products
					</span>
					<h2 className="text-foreground mt-3 text-3xl font-bold tracking-tight md:text-4xl">
						Everything you need, nothing you don't
					</h2>
					<p className="text-muted-foreground mt-4">
						A modular suite of tools designed to grow with your team.
					</p>
				</div>

				<div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
					{products.map((product) => {
						const Icon = product.icon;
						return (
							<Link
								key={product.title}
								to={product.href}
								className="group border-border bg-card hover:border-primary/30 relative flex flex-col rounded-xl border p-6 transition-all hover:-translate-y-0.5 hover:shadow-lg"
							>
								<div
									className={`mb-4 inline-flex h-11 w-11 items-center justify-center rounded-lg ${accentMap[product.accent]}`}
								>
									<Icon className="h-5 w-5" />
								</div>
								<h3 className="text-card-foreground mb-2 text-lg font-semibold">
									{product.title}
								</h3>
								<p className="text-muted-foreground flex-1 text-sm">
									{product.description}
								</p>
								<div className="text-primary mt-5 flex items-center gap-1.5 text-sm font-medium">
									Explore
									<ArrowUpRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
								</div>
							</Link>
						);
					})}
				</div>
			</div>
		</section>
	);
}
