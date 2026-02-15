import Link from "next/link";
import { Button } from "~/components/ui/button";
import { ArrowRight, Github } from "lucide-react";
import RetroGrid from "~/components/magicui/retro-grid";
import WordPullUp from "~/components/magicui/word-pull-up";

export default function Hero() {
  return (
    <section className="bg-background relative flex min-h-[80vh] w-full flex-col items-center justify-center overflow-hidden pt-16 pb-16 md:pt-20 md:pb-20 lg:pt-32 lg:pb-32">
      <RetroGrid />
      <div className="relative z-10 container px-4 md:px-6">
        <div className="flex flex-col items-center gap-8 text-center">
          <div className="animate-fade-in inline-flex items-center rounded-full border border-slate-200 bg-white/80 px-3 py-1 text-sm font-medium text-slate-900 backdrop-blur-sm dark:border-slate-800 dark:bg-slate-900/80 dark:text-slate-50">
            <span className="mr-2 flex h-2 w-2 rounded-full bg-indigo-500"></span>
            v1.0 is now live
          </div>

          <WordPullUp
            words="The Adaptive Enterprise Dashboard for Modern Teams"
            className="max-w-4xl text-4xl font-bold tracking-tight text-slate-900 sm:text-5xl md:text-6xl lg:text-7xl dark:text-slate-50"
          />

          <p className="animate-delay-200 animate-fade-in max-w-[42rem] leading-normal text-slate-500 sm:text-xl sm:leading-8 dark:text-slate-400">
            Bridge the gap between SaaS speed and Enterprise density. NexusOS
            provides a dual-mode interface that adapts to your workflow.
          </p>

          <div className="flex flex-col items-center gap-4 sm:flex-row">
            <Link href="/register">
              <Button
                size="lg"
                className="h-12 rounded-full px-8 shadow-lg shadow-indigo-500/20"
              >
                Get Started
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
                className="h-12 rounded-full px-8 backdrop-blur-sm"
              >
                <Github className="mr-2 h-4 w-4" />
                GitHub
              </Button>
            </Link>
          </div>

          <div className="relative mt-16 aspect-video w-full max-w-5xl overflow-hidden rounded-xl border border-slate-200 bg-slate-100/50 shadow-2xl backdrop-blur-sm dark:border-slate-700 dark:bg-slate-800/50">
            <div className="absolute inset-0 flex items-center justify-center text-slate-400">
              {/* Placeholder for Hero Image/Dashboard Screenshot */}
              <div className="flex flex-col items-center gap-4">
                <span className="text-2xl font-semibold text-slate-600 dark:text-slate-300">
                  NexusOS Interface
                </span>
                <span className="text-sm opacity-60">
                  Interactive Preview Coming Soon
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
