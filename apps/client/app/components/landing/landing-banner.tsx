import { Link } from "react-router";
import { ArrowRight, Palette, Component, Layers } from "lucide-react";

export function LandingBanner() {
  return (
    <section className="border-border border-b py-20 md:py-24">
      <div className="mx-auto max-w-7xl px-6">
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
          {/* Big banner */}
          <Link
            to="/design-system"
            className="group from-primary to-accent text-primary-foreground relative col-span-1 overflow-hidden rounded-2xl bg-gradient-to-br p-8 transition-all hover:shadow-xl lg:col-span-2 lg:p-12"
          >
            <div className="bg-primary-foreground/10 pointer-events-none absolute -top-8 -right-8 h-64 w-64 rounded-full blur-2xl" />
            <div className="bg-primary-foreground/10 pointer-events-none absolute -bottom-12 -left-8 h-72 w-72 rounded-full blur-3xl" />

            <div className="relative">
              <div className="bg-primary-foreground/15 mb-4 inline-flex items-center gap-2 rounded-full px-3 py-1 text-xs font-medium backdrop-blur-sm">
                <Palette className="h-3.5 w-3.5" />
                Design System
              </div>
              <h3 className="mb-3 text-3xl font-bold md:text-4xl">
                A design system you'll actually love using
              </h3>
              <p className="text-primary-foreground/90 mb-6 max-w-md">
                Tokens, components, and patterns crafted for consistency across
                light and dark themes.
              </p>
              <div className="inline-flex items-center gap-2 text-sm font-semibold">
                Explore design tokens
                <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
              </div>
            </div>
          </Link>

          {/* Stacked secondary cards */}
          <div className="flex flex-col gap-6">
            <Link
              to="/components"
              className="group border-border bg-card hover:border-primary/30 flex-1 overflow-hidden rounded-2xl border p-6 transition-all hover:-translate-y-0.5 hover:shadow-lg"
            >
              <Component className="text-primary mb-3 h-7 w-7" />
              <h4 className="text-card-foreground mb-1 font-semibold">
                Components
              </h4>
              <p className="text-muted-foreground mb-3 text-sm">
                90+ ready-to-use components
              </p>
              <div className="text-primary flex items-center gap-1 text-sm font-medium">
                Browse
                <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5" />
              </div>
            </Link>
            <Link
              to="/auth-showcase"
              className="group border-border bg-card hover:border-primary/30 flex-1 overflow-hidden rounded-2xl border p-6 transition-all hover:-translate-y-0.5 hover:shadow-lg"
            >
              <Layers className="text-accent mb-3 h-7 w-7" />
              <h4 className="text-card-foreground mb-1 font-semibold">
                Templates
              </h4>
              <p className="text-muted-foreground mb-3 text-sm">
                Auth flows & error pages
              </p>
              <div className="text-primary flex items-center gap-1 text-sm font-medium">
                View gallery
                <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5" />
              </div>
            </Link>
          </div>
        </div>
      </div>
    </section>
  );
}
