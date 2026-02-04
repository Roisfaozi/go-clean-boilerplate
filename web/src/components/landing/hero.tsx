import Link from "next/link";
import { Button } from "~/components/ui/button";
import { ArrowRight, Github } from "lucide-react";

export default function Hero() {
  return (
    <section className="relative overflow-hidden pt-16 pb-16 md:pt-20 md:pb-20 lg:pt-32 lg:pb-32">
      <div className="container px-4 md:px-6">
        <div className="flex flex-col items-center gap-8 text-center">
          <div className="inline-flex items-center rounded-full border border-slate-200 bg-white px-3 py-1 text-sm font-medium text-slate-900 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-50">
            <span className="mr-2 flex h-2 w-2 rounded-full bg-indigo-500"></span>
            v1.0 is now live
          </div>

          <h1 className="max-w-4xl text-4xl font-bold tracking-tight text-slate-900 sm:text-5xl md:text-6xl lg:text-7xl dark:text-slate-50">
            The Adaptive{" "}
            <span className="text-indigo-600 dark:text-indigo-400">
              Enterprise Dashboard
            </span>{" "}
            for Modern Teams
          </h1>

          <p className="max-w-[42rem] leading-normal text-slate-500 sm:text-xl sm:leading-8 dark:text-slate-400">
            Bridge the gap between SaaS speed and Enterprise density. NexusOS
            provides a dual-mode interface that adapts to your workflow.
          </p>

          <div className="flex flex-col items-center gap-4 sm:flex-row">
            <Link href="/register">
              <Button size="lg" className="h-12 rounded-full px-8">
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
                className="h-12 rounded-full px-8"
              >
                <Github className="mr-2 h-4 w-4" />
                GitHub
              </Button>
            </Link>
          </div>

          <div className="relative mt-16 aspect-video w-full max-w-5xl overflow-hidden rounded-xl border border-slate-200 bg-slate-100 shadow-2xl dark:border-slate-700 dark:bg-slate-800">
            <div className="absolute inset-0 flex items-center justify-center text-slate-400">
              {/* Placeholder for Hero Image/Dashboard Screenshot */}
              <span className="text-lg">Dashboard Preview</span>
            </div>
          </div>
        </div>
      </div>

      {/* Background Gradient */}
      <div className="absolute inset-0 -z-10 h-full w-full bg-white opacity-20 [background:radial-gradient(125%_125%_at_50%_10%,#fff_40%,#6366f1_100%)] dark:bg-slate-950 dark:[background:radial-gradient(125%_125%_at_50%_10%,#020617_40%,#6366f1_100%)]"></div>
    </section>
  );
}
