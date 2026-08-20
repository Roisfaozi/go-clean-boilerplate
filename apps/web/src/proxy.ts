import { createI18nMiddleware } from "next-international/middleware";
import { NextResponse, type NextRequest } from "next/server";

const I18nMiddleware = createI18nMiddleware({
	locales: ["en", "fr"],
	defaultLocale: "en",
});

const locales = ["en", "fr"];

export function proxy(request: NextRequest) {
	const { pathname, search } = request.nextUrl;

	if (request.method === "POST") {
		return I18nMiddleware(request);
	}

	const token = request.cookies.get("access_token")?.value;

	// Helper to extract locale from path safely
	const localeMatch = pathname.match(/^\/([a-z]{2})(?:\/|$)/);
	const detectedLocale =
		localeMatch && locales.includes(localeMatch[1]) ? localeMatch[1] : null;
	const localePrefix = detectedLocale ? `/${detectedLocale}` : "";

	// 1. Protect /dashboard routes (fast path: missing cookie)
	const isDashboardPath =
		pathname.startsWith(`${localePrefix}/dashboard`) ||
		pathname.startsWith("/dashboard");

	if (isDashboardPath && !token) {
		const returnTo = encodeURIComponent(pathname + search);
		const loginUrl = new URL(
			`${localePrefix}/login?returnTo=${returnTo}`,
			request.url,
		);
		return NextResponse.redirect(loginUrl);
	}

	// NOTE: we do NOT bounce logged-in users away from auth pages here.
	// Cookie-presence is not proof of a valid session (tokens can be expired),
	// and bouncing here created an infinite /login <-> /dashboard redirect loop.
	// Auth pages validate the real session server-side (see (auth)/login/page.tsx).

	// 2. Handle Internationalization
	return I18nMiddleware(request);
}

export const config = {
	matcher: [
		"/((?!api|static|.*\\..*|_next|favicon.ico|sitemap.xml|robots.txt).*)",
	],
};
