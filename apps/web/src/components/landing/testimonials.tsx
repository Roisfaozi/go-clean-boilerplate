import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";

const testimonials = [
  {
    name: "Sarah Chen",
    role: "CTO at TechFlow",
    content:
      "NexusOS saved us months of development time. The dual-density feature is a game changer for our logistics dashboard.",
    initials: "SC",
  },
  {
    name: "Michael Ross",
    role: "Indie Hacker",
    content:
      "I launched my SaaS in a weekend. The authentication and billing modules were plug-and-play. Best investment ever.",
    initials: "MR",
  },
  {
    name: "David Kim",
    role: "Senior Engineer",
    content:
      "Finally, a dashboard template that respects TypeScript strict mode and clean architecture. A joy to work with.",
    initials: "DK",
  },
];

export default function Testimonials() {
  return (
    <section className="bg-white py-24 dark:bg-slate-950">
      <div className="container px-4 md:px-6">
        <h2 className="mb-16 text-center text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl dark:text-slate-50">
          Loved by developers & teams
        </h2>

        <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
          {testimonials.map((t) => (
            <div
              key={t.name}
              className="flex flex-col rounded-2xl border border-slate-100 bg-slate-50 p-8 dark:border-slate-800 dark:bg-slate-900"
            >
              <p className="mb-8 text-lg text-slate-600 italic dark:text-slate-300">
                &quot;{t.content}&quot;
              </p>

              <div className="mt-auto flex items-center gap-4">
                <Avatar>
                  <AvatarImage src="" />
                  <AvatarFallback className="bg-indigo-100 text-indigo-700">
                    {t.initials}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <h4 className="font-semibold text-slate-900 dark:text-slate-50">
                    {t.name}
                  </h4>
                  <p className="text-sm text-slate-500 dark:text-slate-400">
                    {t.role}
                  </p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
