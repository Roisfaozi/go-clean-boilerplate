import { Link, useLocation } from "react-router";
import { cn } from "@casbin/ui";
import { useUIStore } from "@/stores";
import {
	LayoutDashboard,
	Users,
	Shield,
	Building2,
	FolderKanban,
	Settings,
	ChevronLeft,
	ChevronRight,
	Hexagon,
	Palette,
	Component,
	KeyRound,
	Lock,
	ChevronDown,
	FileText,
	UserCheck,
	ShieldCheck,
	Upload,
	HeartPulse,
	BarChart3,
	Box,
	Globe,
	LogIn,
	UserPlus,
	KeySquare,
	AlertTriangle,
} from "lucide-react";
import { OrganizationSwitcher } from "@/features/organizations/organization-switcher";
import { Tooltip, TooltipContent, TooltipTrigger } from "@casbin/ui";
import { PresenceAvatars } from "@/components/realtime/presence-avatars";

interface NavItem {
	label: string;
	path: string;
	icon: React.ElementType;
}

interface NavSection {
	label: string;
	items: NavItem[];
	defaultOpen?: boolean;
	subSections?: NavSubSection[];
}

interface NavSubSection {
	label: string;
	icon: React.ElementType;
	items: NavItem[];
}

const navSections: NavSection[] = [
	{
		label: "Dashboard",
		defaultOpen: true,
		items: [{ label: "Overview", path: "/dashboard", icon: LayoutDashboard }],
	},
	{
		label: "Users",
		defaultOpen: true,
		items: [
			{ label: "Users", path: "/users", icon: Users },
			{ label: "Roles", path: "/roles", icon: Shield },
			{ label: "Permissions", path: "/permissions", icon: Lock },
			{
				label: "Role Permissions",
				path: "/roles-permissions",
				icon: ShieldCheck,
			},
			{ label: "Resources", path: "/resources", icon: Box },
			{ label: "Endpoints", path: "/endpoints", icon: Globe },
		],
	},
	{
		label: "Organization",
		defaultOpen: true,
		items: [
			{ label: "Workspace", path: "/workspace", icon: Building2 },
			{ label: "Organizations", path: "/organizations", icon: Building2 },
			{ label: "Members", path: "/workspace", icon: UserCheck },
			{ label: "Projects", path: "/projects", icon: FolderKanban },
		],
	},
	{
		label: "System",
		items: [
			{ label: "Audit Logs", path: "/audit-logs", icon: FileText },
			{ label: "API Docs", path: "/api/v1/docs/index.html", icon: FileText },
			{ label: "System Health", path: "/system-health", icon: HeartPulse },
			{ label: "System Insights", path: "/system-insights", icon: BarChart3 },
			{ label: "Uploads", path: "/uploads", icon: Upload },
			{ label: "Settings", path: "/settings", icon: Settings },
			{ label: "Design System", path: "/design-system", icon: Palette },
			{ label: "Components", path: "/components", icon: Component },
		],
	},
	{
		label: "Showcase",
		items: [
			{ label: "Auth Showcase", path: "/auth-showcase", icon: LogIn },
			{ label: "Error Showcase", path: "/error-showcase", icon: AlertTriangle },
		],
		subSections: [
			{
				label: "Auth Variations",
				icon: LogIn,
				items: [
					{ label: "Login V1", path: "/auth/login-v1", icon: LogIn },
					{ label: "Login V2", path: "/auth/login-v2", icon: LogIn },
					{ label: "Login V3", path: "/auth/login-v3", icon: LogIn },
					{ label: "Register V1", path: "/auth/register-v1", icon: UserPlus },
					{ label: "Register V2", path: "/auth/register-v2", icon: UserPlus },
					{ label: "Register V3", path: "/auth/register-v3", icon: UserPlus },
					{
						label: "Forgot V1",
						path: "/auth/forgot-password-v1",
						icon: KeySquare,
					},
					{
						label: "Forgot V2",
						path: "/auth/forgot-password-v2",
						icon: KeySquare,
					},
					{
						label: "Forgot V3",
						path: "/auth/forgot-password-v3",
						icon: KeySquare,
					},
					{
						label: "Reset V1",
						path: "/auth/reset-password-v1",
						icon: KeyRound,
					},
					{
						label: "Reset V2",
						path: "/auth/reset-password-v2",
						icon: KeyRound,
					},
					{
						label: "Reset V3",
						path: "/auth/reset-password-v3",
						icon: KeyRound,
					},
				],
			},
			{
				label: "Error Pages",
				icon: AlertTriangle,
				items: [
					{ label: "401 V1", path: "/errors/401-v1", icon: AlertTriangle },
					{ label: "401 V2", path: "/errors/401-v2", icon: AlertTriangle },
					{ label: "401 V3", path: "/errors/401-v3", icon: AlertTriangle },
					{ label: "403 V1", path: "/errors/403-v1", icon: AlertTriangle },
					{ label: "403 V2", path: "/errors/403-v2", icon: AlertTriangle },
					{ label: "403 V3", path: "/errors/403-v3", icon: AlertTriangle },
					{ label: "404 V1", path: "/errors/404-v1", icon: AlertTriangle },
					{ label: "404 V2", path: "/errors/404-v2", icon: AlertTriangle },
					{ label: "404 V3", path: "/errors/404-v3", icon: AlertTriangle },
					{ label: "500 V1", path: "/errors/500-v1", icon: AlertTriangle },
					{ label: "500 V2", path: "/errors/500-v2", icon: AlertTriangle },
					{ label: "500 V3", path: "/errors/500-v3", icon: AlertTriangle },
				],
			},
			{
				label: "Auth Standard",
				icon: LogIn,
				items: [
					{ label: "Login", path: "/login", icon: LogIn },
					{ label: "Register", path: "/register", icon: UserPlus },
					{
						label: "Forgot Password",
						path: "/forgot-password",
						icon: KeySquare,
					},
					{ label: "Reset Password", path: "/reset-password", icon: KeyRound },
				],
			},
		],
	},
];

function SidebarSection({
	section,
	collapsed,
	currentPath,
}: {
	section: NavSection;
	collapsed: boolean;
	currentPath: string;
}) {
	const hasActiveChild =
		section.items.some((i) => i.path === currentPath) ||
		(section.subSections?.some((s) =>
			s.items.some((i) => i.path === currentPath),
		) ??
			false);
	const sectionKey = `section:${section.label}`;
	const { sidebarSectionOpen, setSidebarSectionOpen } = useUIStore();
	const stored = sidebarSectionOpen[sectionKey];
	const open = stored ?? section.defaultOpen ?? hasActiveChild;
	const setOpen = (value: boolean) => setSidebarSectionOpen(sectionKey, value);

	if (collapsed) {
		return (
			<div className="space-y-0.5">
				{section.items.map((item) => {
					const isActive = currentPath === item.path;
					return (
						<Tooltip key={item.path + item.label} delayDuration={0}>
							<TooltipTrigger asChild>
								<Link
									to={item.path}
									className={cn(
										"mx-auto flex h-10 w-10 items-center justify-center rounded-md transition-colors duration-150",
										isActive
											? "bg-primary/10 text-primary"
											: "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground",
									)}
								>
									<item.icon className="h-5 w-5" />
								</Link>
							</TooltipTrigger>
							<TooltipContent side="right" sideOffset={8}>
								{item.label}
							</TooltipContent>
						</Tooltip>
					);
				})}
			</div>
		);
	}

	return (
		<div>
			<button
				onClick={() => setOpen(!open)}
				className="text-muted-foreground/70 hover:text-muted-foreground flex w-full items-center justify-between px-3 py-1.5 text-[11px] font-semibold tracking-wider uppercase transition-colors"
			>
				<span>{section.label}</span>
				<ChevronDown
					className={cn(
						"h-3.5 w-3.5 transition-transform duration-200",
						open && "rotate-180",
					)}
				/>
			</button>
			{open && (
				<div className="mt-0.5 space-y-0.5">
					{section.items.map((item) => {
						const isActive = currentPath === item.path;
						return (
							<Link
								key={item.path + item.label}
								to={item.path}
								className={cn(
									"flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors duration-150",
									isActive
										? "bg-primary/10 text-primary"
										: "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground",
								)}
							>
								<item.icon className="h-4.5 w-4.5 shrink-0" />
								<span className="truncate">{item.label}</span>
							</Link>
						);
					})}
					{section.subSections?.map((sub) => (
						<SidebarSubSection
							key={sub.label}
							subSection={sub}
							parentLabel={section.label}
							currentPath={currentPath}
						/>
					))}
				</div>
			)}
		</div>
	);
}

function SidebarSubSection({
	subSection,
	parentLabel,
	currentPath,
}: {
	subSection: NavSubSection;
	parentLabel: string;
	currentPath: string;
}) {
	const hasActiveChild = subSection.items.some((i) => i.path === currentPath);
	const subKey = `sub:${parentLabel}:${subSection.label}`;
	const { sidebarSectionOpen, setSidebarSectionOpen } = useUIStore();
	const stored = sidebarSectionOpen[subKey];
	const open = stored ?? hasActiveChild;
	const setOpen = (value: boolean) => setSidebarSectionOpen(subKey, value);
	const Icon = subSection.icon;

	return (
		<div className="pt-1">
			<button
				onClick={() => setOpen(!open)}
				className={cn(
					"flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors duration-150",
					hasActiveChild
						? "text-sidebar-foreground"
						: "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground",
				)}
			>
				<Icon className="h-4.5 w-4.5 shrink-0" />
				<span className="flex-1 truncate text-left">{subSection.label}</span>
				<ChevronDown
					className={cn(
						"h-3.5 w-3.5 shrink-0 transition-transform duration-200",
						open && "rotate-180",
					)}
				/>
			</button>
			{open && (
				<div className="border-sidebar-border mt-0.5 ml-3 space-y-0.5 border-l pl-3">
					{subSection.items.map((item) => {
						const isActive = currentPath === item.path;
						return (
							<Link
								key={item.path + item.label}
								to={item.path}
								className={cn(
									"flex items-center gap-3 rounded-md px-3 py-1.5 text-sm font-medium transition-colors duration-150",
									isActive
										? "bg-primary/10 text-primary"
										: "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground",
								)}
							>
								<item.icon className="h-4 w-4 shrink-0" />
								<span className="truncate">{item.label}</span>
							</Link>
						);
					})}
				</div>
			)}
		</div>
	);
}

export function AppSidebar() {
	const location = useLocation();
	const { sidebarCollapsed, toggleSidebarCollapse } = useUIStore();

	return (
		<aside
			className={cn(
				"bg-sidebar border-sidebar-border sticky top-0 z-20 flex h-screen flex-col border-r transition-all duration-300",
				sidebarCollapsed ? "w-[68px]" : "w-[240px]",
			)}
		>
			{/* Logo */}
			<div className="border-sidebar-border flex h-14 shrink-0 items-center gap-3 border-b px-4">
				<Hexagon className="text-primary h-7 w-7 shrink-0" />
				{!sidebarCollapsed && (
					<span className="text-sidebar-foreground text-lg font-bold whitespace-nowrap">
						NexusOS
					</span>
				)}
			</div>

			{/* Org Switcher */}
			<div className="border-sidebar-border border-b px-2 py-2">
				<OrganizationSwitcher collapsed={sidebarCollapsed} />
			</div>

			{/* Nav sections */}
			<nav className="flex-1 space-y-4 overflow-y-auto px-2 py-3">
				{navSections.map((section) => (
					<SidebarSection
						key={section.label}
						section={section}
						collapsed={sidebarCollapsed}
						currentPath={location.pathname}
					/>
				))}
			</nav>

			{/* Real-time Presence */}
			{!sidebarCollapsed && (
				<div className="border-sidebar-border bg-sidebar-accent/30 mx-2 mb-4 rounded-lg border p-3">
					<div className="mb-2 flex items-center justify-between">
						<span className="text-muted-foreground text-[10px] font-bold tracking-wider uppercase">
							Who's Online
						</span>
						<div className="bg-success h-1.5 w-1.5 animate-pulse rounded-full" />
					</div>
					<PresenceAvatars max={4} size="sm" showCount />
				</div>
			)}

			{/* Collapse toggle */}
			<div className="border-sidebar-border shrink-0 border-t p-2">
				<button
					onClick={toggleSidebarCollapse}
					className="hover:bg-sidebar-accent text-muted-foreground flex w-full items-center justify-center rounded-md py-2 transition-colors"
				>
					{sidebarCollapsed ? (
						<ChevronRight className="h-4 w-4" />
					) : (
						<ChevronLeft className="h-4 w-4" />
					)}
				</button>
			</div>
		</aside>
	);
}
