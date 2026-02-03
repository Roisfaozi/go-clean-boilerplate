import Link from "next/link";
import { Button } from "~/components/ui/button";
import { ArrowRight, Github } from "lucide-react";

export default function Hero() {
  return (
    <section className="relative overflow-hidden pt-16 md:pt-20 lg:pt-32 pb-16 md:pb-20 lg:pb-32">
      <div className="container px-4 md:px-6">
        <div className="flex flex-col items-center text-center gap-8">
          <div className="inline-flex items-center rounded-full border border-slate-200 bg-white px-3 py-1 text-sm font-medium text-slate-900 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-50">
            <span className="flex h-2 w-2 rounded-full bg-indigo-500 mr-2"></span>
            v1.0 is now live
          </div>
          
          <h1 className="text-4xl font-bold tracking-tight text-slate-900 dark:text-slate-50 sm:text-5xl md:text-6xl lg:text-7xl max-w-4xl">
            The Adaptive <span className="text-indigo-600 dark:text-indigo-400">Enterprise Dashboard</span> for Modern Teams
          </h1>
          
          <p className="max-w-[42rem] leading-normal text-slate-500 dark:text-slate-400 sm:text-xl sm:leading-8">
            Bridge the gap between SaaS speed and Enterprise density. 
            NexusOS provides a dual-mode interface that adapts to your workflow.
          </p>
          
          <div className="flex flex-col sm:flex-row gap-4 items-center">
            <Link href="/register">
              <Button size="lg" className="h-12 px-8 rounded-full">
                Get Started
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
            <Link href="https://github.com/Roisfaozi/go-clean-boilerplate" target="_blank">
              <Button variant="outline" size="lg" className="h-12 px-8 rounded-full">
                <Github className="mr-2 h-4 w-4" />
                GitHub
              </Button>
            </Link>
          </div>

          <div className="mt-16 relative w-full max-w-5xl aspect-video rounded-xl bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-2xl overflow-hidden">
            <div className="absolute inset-0 flex items-center justify-center text-slate-400">
              {/* Placeholder for Hero Image/Dashboard Screenshot */}
              <span className="text-lg">Dashboard Preview</span>
            </div>
          </div>
        </div>
      </div>
      
      {/* Background Gradient */}
      <div className="absolute inset-0 -z-10 h-full w-full bg-white dark:bg-slate-950 [background:radial-gradient(125%_125%_at_50%_10%,#fff_40%,#6366f1_100%)] dark:[background:radial-gradient(125%_125%_at_50%_10%,#020617_40%,#6366f1_100%)] opacity-20"></div>
    </section>
  );
}
