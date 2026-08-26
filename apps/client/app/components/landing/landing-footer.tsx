import { Link } from "react-router";
import { Hexagon, Globe, Mail, MessageSquare } from "lucide-react";

const sections = [
	{
		title: "Product",
		links: [
			{ label: "Features", href: "#features" },
			{ label: "Components", href: "/components" },
			{ label: "Design System", href: "/design-system" },
			{ label: "Templates", href: "/auth-showcase" },
		],
	},
	{
		title: "Resources",
		links: [
			{ label: "Documentation", href: "#" },
			{ label: "API Reference", href: "#" },
			{ label: "Changelog", href: "#" },
			{ label: "Status", href: "/system-health" },
		],
	},
	{
		title: "Company",
		links: [
			{ label: "About", href: "#" },
			{ label: "Blog", href: "#" },
			{ label: "Careers", href: "#" },
			{ label: "Contact", href: "#" },
		],
	},
];

export function LandingFooter() {
	return (
		<footer className="bg-surface/50">
			<div className="mx-auto max-w-7xl px-6 py-16">
				<div className="grid grid-cols-2 gap-10 md:grid-cols-5">
					<div className="col-span-2">
						<Link to="/" className="flex items-center gap-2">
							<Hexagon className="text-primary h-7 w-7" />
							<span className="text-lg font-bold tracking-tight">NexusOS</span>
						</Link>
						<p className="text-muted-foreground mt-4 max-w-xs text-sm">
							The operating system for modern teams. Build, ship, and scale with
							confidence.
						</p>
						<div className="mt-6 flex items-center gap-3">
							{[Globe, Mail, MessageSquare].map((Icon, i) => (
								<a
									key={i}
									href="/"
									className="border-border bg-background text-muted-foreground hover:border-primary/30 hover:text-primary flex h-9 w-9 items-center justify-center rounded-md border transition-colors"
								>
									<Icon className="h-4 w-4" />
								</a>
							))}
						</div>
					</div>

					{sections.map((section) => (
						<div key={section.title}>
							<h4 className="text-foreground mb-4 text-sm font-semibold">
								{section.title}
							</h4>
							<ul className="space-y-3">
								{section.links.map((link) => (
									<li key={link.label}>
										<a
											href={link.href}
											className="text-muted-foreground hover:text-foreground text-sm transition-colors"
										>
											{link.label}
										</a>
									</li>
								))}
							</ul>
						</div>
					))}
				</div>

				<div className="border-border mt-12 flex flex-col items-center justify-between gap-4 border-t pt-8 md:flex-row">
					<p className="text-muted-foreground text-xs">
						© {new Date().getFullYear()} NexusOS. All rights reserved.
					</p>
					<div className="text-muted-foreground flex items-center gap-6 text-xs">
						<a href="/privacy" className="hover:text-foreground">
							Privacy
						</a>
						<a href="/terms" className="hover:text-foreground">
							Terms
						</a>
						<a href="/cookies" className="hover:text-foreground">
							Cookies
						</a>
					</div>
				</div>
			</div>
		</footer>
	);
}
