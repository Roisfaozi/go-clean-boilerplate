import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
} from "react-router";
import { useEffect, useRef, useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

import { useUIStore } from "@/stores";
import { useAuthStore } from "@/stores/auth-store";
import { authApi } from "@/lib/api/auth";
import { Toaster, Sonner, TooltipProvider } from "@casbin/ui";

import type { Route } from "./+types/root";
import "./app.css";

export const links: Route.LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap",
  },
];

function ThemeInitializer({ children }: { children: React.ReactNode }) {
  const { theme, density } = useUIStore();

  useEffect(() => {
    document.documentElement.classList.toggle("dark", theme === "dark");
    document.documentElement.classList.toggle(
      "density-compact",
      density === "compact",
    );
  }, [theme, density]);

  return <>{children}</>;
}

function SessionInitializer({ children }: { children: React.ReactNode }) {
  const initialized = useAuthStore((s) => s.initialized);
  const setUser = useAuthStore((s) => s.setUser);
  const setInitialized = useAuthStore((s) => s.setInitialized);
  const hasBootstrapped = useRef(false);

  useEffect(() => {
    if (hasBootstrapped.current) {
      return;
    }

    hasBootstrapped.current = true;

    let mounted = true;

    const bootstrapSession = async () => {
      try {
        const user = await authApi.getCurrentUser();
        if (mounted) {
          setUser(user);
        }
      } catch {
        // No active session or session expired.
      } finally {
        if (mounted) {
          setInitialized(true);
        }
      }
    };

    bootstrapSession();

    return () => {
      mounted = false;
    };
  }, [setInitialized, setUser]);

  if (!initialized) {
    return (
      <div className="bg-background text-foreground flex min-h-screen items-center justify-center">
        <span>Memeriksa sesi...</span>
      </div>
    );
  }

  return <>{children}</>;
}

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App() {
  const [queryClient] = useState(() => new QueryClient());

  return (
    <QueryClientProvider client={queryClient}>
      <SessionInitializer>
        <TooltipProvider>
          <ThemeInitializer>
            <Outlet />
            <Toaster />
            <Sonner />
          </ThemeInitializer>
        </TooltipProvider>
      </SessionInitializer>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  let message = "Oops!";
  let details = "An unexpected error occurred.";
  let stack: string | undefined;

  if (isRouteErrorResponse(error)) {
    message = error.status === 404 ? "404" : "Error";
    details =
      error.status === 404
        ? "The requested page could not be found."
        : error.statusText || details;
  } else if (import.meta.env.DEV && error && error instanceof Error) {
    details = error.message;
    stack = error.stack;
  }

  return (
    <main className="container mx-auto p-4 pt-16">
      <h1>{message}</h1>
      <p>{details}</p>
      {stack && (
        <pre className="w-full overflow-x-auto p-4">
          <code>{stack}</code>
        </pre>
      )}
    </main>
  );
}
