import Link from "next/link";
import { Button } from "~/components/ui/button";
import { Github } from "lucide-react";

export default function OpenSource() {
  return (
    <section className="py-24 bg-white dark:bg-slate-950 border-t border-slate-200 dark:border-slate-800">
      <div className="container px-4 md:px-6">
        <div className="flex flex-col items-center justify-center text-center space-y-8">
          <h2 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-slate-50 sm:text-4xl">
            Open Source & Community
          </h2>
          <p className="max-w-[600px] text-lg text-slate-500 dark:text-slate-400">
            NexusOS is built on top of open standards. Join our community to contribute, report bugs, or request features.
          </p>
          <Link href="https://github.com/Roisfaozi/go-clean-boilerplate" target="_blank">
            <Button size="lg" variant="outline" className="h-12 px-8 gap-2">
              <Github className="h-5 w-5" />
              Star on GitHub
            </Button>
          </Link>
        </div>
      </div>
    </section>
  );
}