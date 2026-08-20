import { Link } from "react-router";
import { PageHeader } from "@/components/layout/page-header";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@casbin/ui";
import { Badge } from "@casbin/ui";
import { NexusButton } from "@casbin/ui";
import {
	ExternalLink,
	Layout,
	Sparkles,
	Moon,
	LogIn,
	UserPlus,
	KeyRound,
	ShieldCheck,
} from "lucide-react";

interface Variant {
	id: string;
	title: string;
	description: string;
	tags: string[];
	path: string;
	badge: { icon: React.ElementType; label: string; className: string };
}

const loginVariants: Variant[] = [
	{
		id: "lv1",
		title: "Split Screen",
		description:
			"Layout dua kolom dengan branding panel animasi di kiri dan form login di kanan. Cocok untuk enterprise apps.",
		tags: ["Split Layout", "Animated Shapes", "Feature List"],
		path: "/auth/login-v1",
		badge: {
			icon: Layout,
			label: "Split",
			className: "bg-primary text-primary-foreground",
		},
	},
	{
		id: "lv2",
		title: "Glassmorphism Card",
		description:
			"Card transparan dengan efek blur di atas background gradient animasi. Modern dan elegan untuk SaaS products.",
		tags: ["Centered Card", "Glassmorphism", "Gradient BG"],
		path: "/auth/login-v2",
		badge: {
			icon: Sparkles,
			label: "Glass",
			className: "bg-accent text-accent-foreground",
		},
	},
	{
		id: "lv3",
		title: "Dark Immersive",
		description:
			"Full-screen dark theme dengan grid pattern, neon glow, testimonial, dan stats.",
		tags: ["Dark Theme", "Grid Pattern", "Testimonial"],
		path: "/auth/login-v3",
		badge: {
			icon: Moon,
			label: "Dark",
			className: "bg-foreground text-background",
		},
	},
];

const registerVariants: Variant[] = [
	{
		id: "rv1",
		title: "Split Screen",
		description:
			"Register form di kanan dengan branding panel animasi di kiri. Termasuk password strength indicator.",
		tags: ["Split Layout", "Password Strength", "Animated Shapes"],
		path: "/auth/register-v1",
		badge: {
			icon: Layout,
			label: "Split",
			className: "bg-primary text-primary-foreground",
		},
	},
	{
		id: "rv2",
		title: "Glassmorphism Card",
		description:
			"Register card transparan dengan glassmorphism effect di atas animated gradient background.",
		tags: ["Centered Card", "Glassmorphism", "Gradient BG"],
		path: "/auth/register-v2",
		badge: {
			icon: Sparkles,
			label: "Glass",
			className: "bg-accent text-accent-foreground",
		},
	},
	{
		id: "rv3",
		title: "Dark Immersive",
		description:
			"Full-screen dark register dengan grid pattern, neon glow, stats, dan testimonial.",
		tags: ["Dark Theme", "Grid Pattern", "Neon Glow"],
		path: "/auth/register-v3",
		badge: {
			icon: Moon,
			label: "Dark",
			className: "bg-foreground text-background",
		},
	},
];

const forgotVariants: Variant[] = [
	{
		id: "fv1",
		title: "Split Screen",
		description:
			"Layout split dengan branding panel kiri berisi step-by-step recovery flow dan form di kanan.",
		tags: ["Split Layout", "Step Guide", "Animated Rings"],
		path: "/auth/forgot-password-v1",
		badge: {
			icon: Layout,
			label: "Split",
			className: "bg-primary text-primary-foreground",
		},
	},
	{
		id: "fv2",
		title: "Glassmorphism Card",
		description:
			"Forgot password card dengan glassmorphism di atas rotating gradient background.",
		tags: ["Centered Card", "Glassmorphism", "Gradient BG"],
		path: "/auth/forgot-password-v2",
		badge: {
			icon: Sparkles,
			label: "Glass",
			className: "bg-accent text-accent-foreground",
		},
	},
	{
		id: "fv3",
		title: "Dark Immersive",
		description:
			"Dark mode forgot password dengan grid pattern dan neon glow effect.",
		tags: ["Dark Theme", "Grid Pattern", "Neon Glow"],
		path: "/auth/forgot-password-v3",
		badge: {
			icon: Moon,
			label: "Dark",
			className: "bg-foreground text-background",
		},
	},
];

const resetVariants: Variant[] = [
	{
		id: "rpv1",
		title: "Split Screen",
		description:
			"Reset password split layout dengan shield branding dan password strength checklist.",
		tags: ["Split Layout", "Strength Check", "Animated Rings"],
		path: "/auth/reset-password-v1",
		badge: {
			icon: Layout,
			label: "Split",
			className: "bg-primary text-primary-foreground",
		},
	},
	{
		id: "rpv2",
		title: "Glassmorphism Card",
		description:
			"Reset password glassmorphism card dengan strength bar visual dan pill indicators.",
		tags: ["Centered Card", "Strength Bar", "Glassmorphism"],
		path: "/auth/reset-password-v2",
		badge: {
			icon: Sparkles,
			label: "Glass",
			className: "bg-accent text-accent-foreground",
		},
	},
	{
		id: "rpv3",
		title: "Dark Immersive",
		description:
			"Dark mode reset password dengan grid pattern, strength pills, dan neon glow.",
		tags: ["Dark Theme", "Grid Pattern", "Strength Pills"],
		path: "/auth/reset-password-v3",
		badge: {
			icon: Moon,
			label: "Dark",
			className: "bg-foreground text-background",
		},
	},
];

function VariantCard({ variant, index }: { variant: Variant; index: number }) {
	const BadgeIcon = variant.badge.icon;
	return (
		<Card className="group border-border hover:border-primary/30 transition-all duration-300 hover:shadow-lg">
			<CardHeader className="space-y-3">
				<div className="flex items-center justify-between">
					<Badge variant="outline" className="text-xs">
						Variant {index + 1}
					</Badge>
					<Badge className={`${variant.badge.className} gap-1 text-xs`}>
						<BadgeIcon className="h-3 w-3" /> {variant.badge.label}
					</Badge>
				</div>
				<CardTitle className="text-lg">{variant.title}</CardTitle>
				<CardDescription className="text-sm leading-relaxed">
					{variant.description}
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="flex flex-wrap gap-1.5">
					{variant.tags.map((tag) => (
						<Badge
							key={tag}
							variant="secondary"
							className="text-xs font-normal"
						>
							{tag}
						</Badge>
					))}
				</div>
				<Link to={variant.path}>
					<NexusButton variant="outline" className="w-full gap-2">
						<ExternalLink className="h-4 w-4" /> Lihat Preview
					</NexusButton>
				</Link>
			</CardContent>
		</Card>
	);
}

function Section({
	icon: Icon,
	title,
	variants,
}: {
	icon: React.ElementType;
	title: string;
	variants: Variant[];
}) {
	return (
		<div className="space-y-4">
			<div className="flex items-center gap-2">
				<Icon className="text-primary h-5 w-5" />
				<h2 className="text-foreground text-xl font-semibold">{title}</h2>
			</div>
			<div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
				{variants.map((v, i) => (
					<VariantCard key={v.id} variant={v} index={i} />
				))}
			</div>
		</div>
	);
}

export default function AuthShowcasePage() {
	return (
		<div className="space-y-8">
			<PageHeader
				title="Auth Page Variations"
				description="Koleksi variasi halaman auth dengan layout dan style yang berbeda"
			/>
			<Section icon={LogIn} title="Login Pages" variants={loginVariants} />
			<Section
				icon={UserPlus}
				title="Register Pages"
				variants={registerVariants}
			/>
			<Section
				icon={KeyRound}
				title="Forgot Password Pages"
				variants={forgotVariants}
			/>
			<Section
				icon={ShieldCheck}
				title="Reset Password Pages"
				variants={resetVariants}
			/>
		</div>
	);
}
