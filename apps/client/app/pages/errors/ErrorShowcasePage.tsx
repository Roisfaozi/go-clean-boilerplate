import { useNavigate } from "react-router";
import { PageHeader } from "@/components/layout/page-header";
import {
	NexusCard,
	NexusCardHeader,
	NexusCardTitle,
	NexusCardDescription,
	NexusCardContent,
} from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import {
	Lock,
	ShieldOff,
	Search,
	ServerCrash,
	ExternalLink,
} from "lucide-react";

const sections = [
	{
		title: "401 — Unauthorized",
		icon: Lock,
		color: "text-warning",
		variants: [
			{ label: "V1 — Split Screen", path: "/errors/401-v1" },
			{ label: "V2 — Glassmorphism", path: "/errors/401-v2" },
			{ label: "V3 — Dark Immersive", path: "/errors/401-v3" },
		],
	},
	{
		title: "403 — Forbidden",
		icon: ShieldOff,
		color: "text-danger",
		variants: [
			{ label: "V1 — Split Screen", path: "/errors/403-v1" },
			{ label: "V2 — Glassmorphism", path: "/errors/403-v2" },
			{ label: "V3 — Dark Immersive", path: "/errors/403-v3" },
		],
	},
	{
		title: "404 — Not Found",
		icon: Search,
		color: "text-primary",
		variants: [
			{ label: "V1 — Split Screen", path: "/errors/404-v1" },
			{ label: "V2 — Glassmorphism", path: "/errors/404-v2" },
			{ label: "V3 — Dark Immersive", path: "/errors/404-v3" },
		],
	},
	{
		title: "500 — Server Error",
		icon: ServerCrash,
		color: "text-destructive",
		variants: [
			{ label: "V1 — Split Screen", path: "/errors/500-v1" },
			{ label: "V2 — Glassmorphism", path: "/errors/500-v2" },
			{ label: "V3 — Dark Immersive", path: "/errors/500-v3" },
		],
	},
];

export default function ErrorShowcasePage() {
	const navigate = useNavigate();

	return (
		<div className="max-w-5xl space-y-8">
			<PageHeader
				title="Error & Empty States"
				description="Showcase of error state pages with 3 visual variations each."
			/>

			<div className="grid grid-cols-1 gap-6 md:grid-cols-2">
				{sections.map((section) => (
					<NexusCard key={section.title}>
						<NexusCardHeader>
							<div className="flex items-center gap-3">
								<div className="bg-muted flex h-10 w-10 items-center justify-center rounded-lg">
									<section.icon className={`h-5 w-5 ${section.color}`} />
								</div>
								<div>
									<NexusCardTitle className="text-base">
										{section.title}
									</NexusCardTitle>
									<NexusCardDescription>
										3 visual variations
									</NexusCardDescription>
								</div>
							</div>
						</NexusCardHeader>
						<NexusCardContent>
							<div className="flex flex-col gap-2">
								{section.variants.map((v) => (
									<NexusButton
										key={v.path}
										variant="outline"
										className="justify-between"
										onClick={() => navigate(v.path)}
									>
										{v.label}
										<ExternalLink className="text-muted-foreground h-3.5 w-3.5" />
									</NexusButton>
								))}
							</div>
						</NexusCardContent>
					</NexusCard>
				))}
			</div>
		</div>
	);
}
