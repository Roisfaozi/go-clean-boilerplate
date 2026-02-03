import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";

const testimonials = [
  {
    name: "Sarah Chen",
    role: "CTO at TechFlow",
    content: "NexusOS saved us months of development time. The dual-density feature is a game changer for our logistics dashboard.",
    initials: "SC",
  },
  {
    name: "Michael Ross",
    role: "Indie Hacker",
    content: "I launched my SaaS in a weekend. The authentication and billing modules were plug-and-play. Best investment ever.",
    initials: "MR",
  },
  {
    name: "David Kim",
    role: "Senior Engineer",
    content: "Finally, a dashboard template that respects TypeScript strict mode and clean architecture. A joy to work with.",
    initials: "DK",
  },
];

export default function Testimonials() {
  return (
    <section className="py-24 bg-white dark:bg-slate-950">
      <div className="container px-4 md:px-6">
        <h2 className="text-3xl font-bold tracking-tight text-center text-slate-900 dark:text-slate-50 mb-16 sm:text-4xl">
          Loved by developers & teams
        </h2>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {testimonials.map((t) => (
            <div key={t.name} className="flex flex-col p-8 rounded-2xl bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-800">
              <p className="text-lg text-slate-600 dark:text-slate-300 mb-8 italic">"{t.content}"</p>
              
              <div className="mt-auto flex items-center gap-4">
                <Avatar>
                  <AvatarImage src="" />
                  <AvatarFallback className="bg-indigo-100 text-indigo-700">{t.initials}</AvatarFallback>
                </Avatar>
                <div>
                  <h4 className="font-semibold text-slate-900 dark:text-slate-50">{t.name}</h4>
                  <p className="text-sm text-slate-500 dark:text-slate-400">{t.role}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}