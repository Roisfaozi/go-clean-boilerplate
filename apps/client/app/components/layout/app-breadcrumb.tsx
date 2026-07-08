import { useLocation, Link } from "react-router";
import {
	Breadcrumb,
	BreadcrumbItem,
	BreadcrumbLink,
	BreadcrumbList,
	BreadcrumbPage,
	BreadcrumbSeparator,
} from "@casbin/ui";
import { Fragment } from "react";

const routeLabels: Record<string, string> = {
	"": "Dashboard",
	users: "Users",
	roles: "Roles",
	permissions: "Permissions",
	"access-rights": "Access Rights",
	organizations: "Organizations",
	projects: "Projects",
	"audit-logs": "Audit Logs",
	settings: "Settings",
	"design-system": "Design System",
	components: "Components",
	buttons: "Buttons",
	badges: "Badges",
	forms: "Forms",
	cards: "Cards",
	feedback: "Feedback",
	"data-display": "Data Display",
	overlays: "Overlays",
	navigation: "Navigation",
	charts: "Charts",
	realtime: "Realtime",
};

export function AppBreadcrumb() {
	const location = useLocation();
	const segments = location.pathname.split("/").filter(Boolean);

	if (segments.length === 0) return null;

	return (
		<Breadcrumb className="mb-4">
			<BreadcrumbList>
				<BreadcrumbItem>
					<BreadcrumbLink asChild>
						<Link to="/">Dashboard</Link>
					</BreadcrumbLink>
				</BreadcrumbItem>
				{segments.map((seg, i) => {
					const path = "/" + segments.slice(0, i + 1).join("/");
					const label = routeLabels[seg] || seg;
					const isLast = i === segments.length - 1;

					return (
						<Fragment key={path}>
							<BreadcrumbSeparator />
							<BreadcrumbItem>
								{isLast ? (
									<BreadcrumbPage>{label}</BreadcrumbPage>
								) : (
									<BreadcrumbLink asChild>
										<Link to={path}>{label}</Link>
									</BreadcrumbLink>
								)}
							</BreadcrumbItem>
						</Fragment>
					);
				})}
			</BreadcrumbList>
		</Breadcrumb>
	);
}
