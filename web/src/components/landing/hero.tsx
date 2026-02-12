"use client";

import Link from "next/link";
import { Button } from "~/components/ui/button";
import { ArrowRight, Github } from "lucide-react";
import RetroGrid from "~/components/magicui/retro-grid";
import WordPullUp from "~/components/magicui/word-pull-up";
import { useAuthStore } from "~/stores/use-auth-store";

export default function Hero() {
  const { user } = useAuthStore();

  return (
    <section className="bg-background relative flex min-h-screen w-full flex-col items-center justify-center overflow-hidden pt-20 pb-32">
      <RetroGrid />
      <div className="relative z-10 container px-4 md:px-6">
        <div className="flex flex-col items-center gap-8 text-center">
          <div className="animate-fade-in inline-flex items-center rounded-full border border-primary/20 bg-primary/5 px-3 py-1 text-xs font-bold tracking-wider text-primary uppercase backdrop-blur-sm">
            <span className="mr-2 flex h-2 w-2 rounded-full bg-primary animate-pulse"></span>
            NexusOS Enterprise v1.0
          </div>

          <WordPullUp
            words="Build Scalable Apps with Enterprise Foundation."
            className="max-w-4xl text-4xl font-extrabold tracking-tight text-slate-900 sm:text-5xl md:text-6xl lg:text-7xl dark:text-slate-50"
          />

          <p className="animate-delay-200 animate-fade-in max-w-[48rem] text-lg text-slate-500 sm:text-xl dark:text-slate-400">
            The ultimate Go + Next.js boilerplate. 
            Granular Casbin RBAC, Multi-tenancy, Real-time Presence, and Modular Audit Logging—everything you need to ship enterprise SaaS in days.
          </p>

          <div className="flex flex-col items-center gap-4 sm:flex-row">
            <Link href={user ? "/dashboard" : "/register"}>
              <Button
                size="lg"
                className="h-12 rounded-xl px-8 text-base shadow-xl shadow-primary/20 transition-all hover:scale-105"
              >
                {user ? "Go to Dashboard" : "Get Started Now"}
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
            <Link
              href="https://github.com/Roisfaozi/go-clean-boilerplate"
              target="_blank"
            >
              <Button
                variant="outline"
                size="lg"
                className="h-12 rounded-xl px-8 text-base backdrop-blur-sm transition-all hover:bg-muted"
              >
                <Github className="mr-2 h-4 w-4" />
                Explore Source
              </Button>
            </Link>
          </div>

          {/* Feature Showcase Box */}
          <div className="relative mt-20 w-full max-w-5xl rounded-2xl border bg-card/50 p-2 shadow-2xl backdrop-blur-sm overflow-hidden">
            <div className="absolute inset-0 bg-linear-to-tr from-primary/5 to-transparent" />
            <div className="bg-background relative rounded-xl border border-dashed p-12 lg:p-24">
               <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
                  <div className="flex flex-col items-center gap-2">
                    <span className="text-3xl font-bold">100%</span>
                    <span className="text-muted-foreground text-xs uppercase font-semibold">TDD Coverage</span>
                  </div>
                  <div className="flex flex-col items-center gap-2 border-l border-dashed pl-8">
                    <span className="text-3xl font-bold">Go 1.25</span>
                    <span className="text-muted-foreground text-xs uppercase font-semibold">Backend Engine</span>
                  </div>
                  <div className="flex flex-col items-center gap-2 border-l border-dashed pl-8">
                    <span className="text-3xl font-bold">Next.js 16</span>
                    <span className="text-muted-foreground text-xs uppercase font-semibold">App Router</span>
                  </div>
                  <div className="flex flex-col items-center gap-2 border-l border-dashed pl-8">
                    <span className="text-3xl font-bold">Casbin</span>
                    <span className="text-muted-foreground text-xs uppercase font-semibold">AuthZ Standard</span>
                  </div>
               </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
