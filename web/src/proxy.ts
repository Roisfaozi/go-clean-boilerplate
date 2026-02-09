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

  // 1. Protect /dashboard routes
  const isDashboardPath =
    pathname.startsWith(`${localePrefix}/dashboard`) ||
    pathname.startsWith("/dashboard");
  const isAuthPath =
    pathname.match(/^\/([a-z]{2})\/(login|register)(?:\/|$)/) ||
    pathname.startsWith("/login") ||
    pathname.startsWith("/register");

  if (isDashboardPath) {
    if (!token) {
      const returnTo = encodeURIComponent(pathname + search);
      const loginUrl = new URL(
        `${localePrefix}/login?returnTo=${returnTo}`,
        request.url
      );
      return NextResponse.redirect(loginUrl);
    }
  }

  // 2. Redirect logged-in users away from auth pages
  if (isAuthPath && token) {
    return NextResponse.redirect(
      new URL(`${localePrefix}/dashboard`, request.url)
    );
  }

  // 3. Handle Internationalization
  return I18nMiddleware(request);
}

export const config = {
  matcher: [
    "/((?!api|static|.*\\..*|_next|favicon.ico|sitemap.xml|robots.txt).*)",
  ],
};
