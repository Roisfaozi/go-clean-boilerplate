import { redirect } from "next/navigation";
import { getCurrentSession } from "~/lib/server/auth/session";
import { organizationsApi } from "~/lib/api/organizations";

// Better way: import usePresence in a dedicated client component or just here if we make this function client
// Let's create a Client wrapper for the layout logic
import { DashboardLayoutClient } from "./layout-client";
import { DashboardShellProvider } from "./_components/dashboard-shell-context";

const locales = ["en", "fr"];

export default async function DashboardLayout({
	children,
	params,
}: {
	children: React.ReactNode;
	params: Promise<{ locale?: string }>;
}) {
	// 1. Validate the auth session on the server. Expired or missing tokens
	//    redirect to the login page before any dashboard shell renders.
	const { session } = await getCurrentSession();
	if (!session) {
		const { locale } = await params;
		const detectedLocale = locale && locales.includes(locale) ? locale : "en";
		redirect(
			`/${detectedLocale}/login?returnTo=${encodeURIComponent(`/${detectedLocale}/dashboard`)}`,
		);
	}

	// 2. Fetch organizations on Server (Critical for Navigation/Switcher)
	let initialOrgs = undefined;
	try {
		const resp = await organizationsApi.getMyOrganizations();
		initialOrgs = resp.data?.organizations;
	} catch (error) {
		console.error("Failed to fetch initial orgs on server", error);
	}

	return (
		<DashboardShellProvider initialData={initialOrgs}>
			<DashboardLayoutClient>{children}</DashboardLayoutClient>
		</DashboardShellProvider>
	);
}