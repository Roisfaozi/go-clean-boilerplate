import { createI18nMiddleware } from "next-international/middleware";
import { NextResponse, type NextRequest } from "next/server";

const I18nMiddleware = createI18nMiddleware({
  locales: ["en", "fr"],
  defaultLocale: "en",
});

export function proxy(request: NextRequest) {
  const token = request.cookies.get("access_token")?.value;
  const { pathname, search } = request.nextUrl;

  // 1. Protect /dashboard routes
  const isDashboardPath =
    pathname.match(/^\/([a-z]{2})\/dashboard/) ||
    pathname.startsWith("/dashboard");
  const isAuthPath =
    pathname.match(/^\/([a-z]{2})\/(login|register)/) ||
    pathname.startsWith("/login") ||
    pathname.startsWith("/register");

  if (isDashboardPath) {
    if (!token) {
      const returnTo = encodeURIComponent(pathname + search);
      const localeMatch = pathname.match(/^\/([a-z]{2})/);
      const localePrefix = localeMatch ? `/${localeMatch[1]}` : "";
      const loginUrl = new URL(
        `${localePrefix}/login?returnTo=${returnTo}`,
        request.url
      );
      return NextResponse.redirect(loginUrl);
    }
  }

  // 2. Redirect logged-in users away from auth pages
  if (isAuthPath && token) {
    const localeMatch = pathname.match(/^\/([a-z]{2})/);
    const localePrefix = localeMatch ? `/${localeMatch[1]}` : "";
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
