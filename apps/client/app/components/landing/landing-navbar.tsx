import { Link } from "react-router";
import { Hexagon, Moon, Sun, Menu, X } from "lucide-react";
import { useState } from "react";
import { useUIStore } from "@/stores";
import { cn } from "@casbin/ui";

const navItems = [
  { label: "Features", href: "#features" },
  { label: "Modules", href: "#modules" },
  { label: "Testimonials", href: "#testimonials" },
  { label: "Pricing", href: "#pricing" },
];

export function LandingNavbar() {
  const { theme, setTheme } = useUIStore();
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <header className="border-border/60 bg-background/80 sticky top-0 z-50 w-full border-b backdrop-blur-lg">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
        <Link to="/" className="flex items-center gap-2">
          <Hexagon className="text-primary h-7 w-7" />
          <span className="text-lg font-bold tracking-tight">NexusOS</span>
        </Link>

        <nav className="hidden items-center gap-8 md:flex">
          {navItems.map((item) => (
            <a
              key={item.href}
              href={item.href}
              className="text-muted-foreground hover:text-foreground text-sm font-medium transition-colors"
            >
              {item.label}
            </a>
          ))}
        </nav>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="text-muted-foreground hover:bg-muted hover:text-foreground flex h-9 w-9 items-center justify-center rounded-md transition-colors"
            aria-label="Toggle theme"
          >
            {theme === "dark" ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
          </button>
          <Link
            to="/login"
            className="text-muted-foreground hover:text-foreground hidden h-9 items-center rounded-md px-4 text-sm font-medium transition-colors md:inline-flex"
          >
            Sign in
          </Link>
          <Link
            to="/register"
            className="bg-primary text-primary-foreground hover:bg-primary-hover hidden h-9 items-center rounded-md px-4 text-sm font-medium shadow-sm transition-colors md:inline-flex"
          >
            Get started
          </Link>
          <button
            onClick={() => setMobileOpen(!mobileOpen)}
            className="text-muted-foreground flex h-9 w-9 items-center justify-center rounded-md md:hidden"
            aria-label="Toggle menu"
          >
            {mobileOpen ? (
              <X className="h-5 w-5" />
            ) : (
              <Menu className="h-5 w-5" />
            )}
          </button>
        </div>
      </div>

      {mobileOpen && (
        <div className="border-border bg-background border-t md:hidden">
          <div className="mx-auto flex max-w-7xl flex-col gap-1 px-6 py-4">
            {navItems.map((item) => (
              <a
                key={item.href}
                href={item.href}
                onClick={() => setMobileOpen(false)}
                className="text-muted-foreground hover:bg-muted hover:text-foreground rounded-md px-3 py-2 text-sm font-medium"
              >
                {item.label}
              </a>
            ))}
            <div className="border-border mt-2 flex gap-2 border-t pt-3">
              <Link
                to="/login"
                className="border-border flex-1 rounded-md border px-3 py-2 text-center text-sm font-medium"
              >
                Sign in
              </Link>
              <Link
                to="/register"
                className={cn(
                  "bg-primary text-primary-foreground flex-1 rounded-md px-3 py-2 text-center text-sm font-medium",
                )}
              >
                Get started
              </Link>
            </div>
          </div>
        </div>
      )}
    </header>
  );
}
