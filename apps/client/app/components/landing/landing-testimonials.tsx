import { Quote } from "lucide-react";

const testimonials = [
  {
    quote:
      "Dengan NexusOS, kami memangkas waktu setup admin platform secara drastis. Semua terasa lebih konsisten dari sisi access, UI, dan operasional.",
    name: "Rina Hartono",
    role: "Head of Product",
    company: "Sentuh Digital",
    initials: "RH",
  },
  {
    quote:
      "Yang paling terasa adalah clarity. Design system, admin flow, dan struktur operasionalnya sangat membantu tim berkembang tanpa chaos.",
    name: "Daniel Kurniawan",
    role: "Engineering Manager",
    company: "Atlas Group",
    initials: "DK",
  },
  {
    quote:
      "Bukan cuma dashboard yang rapi. Governance, auditability, dan readiness untuk scale benar-benar terasa sejak awal.",
    name: "Maya Pranoto",
    role: "CTO",
    company: "Orion Ops",
    initials: "MP",
  },
];

export function LandingTestimonials() {
  return (
    <section id="testimonials" className="border-b border-border bg-surface py-20 md:py-24">
      <div className="mx-auto max-w-7xl px-layout">
        <div className="mx-auto mb-14 max-w-2xl text-center">
          <span className="text-caption font-semibold uppercase tracking-wider text-primary">
            Testimonials
          </span>
          <h2 className="mt-3 text-3xl font-bold tracking-tight text-foreground md:text-4xl">
            Loved by teams building serious products
          </h2>
          <p className="mt-4 text-body-lg text-muted-foreground">
            Bukti dari tim yang sudah membangun di atas NexusOS.
          </p>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {testimonials.map((t) => (
            <figure
              key={t.name}
              className="relative flex flex-col rounded-xl border border-border bg-card p-6"
            >
              <Quote className="mb-4 h-6 w-6 text-primary/40" />
              <blockquote className="mb-6 flex-1 text-sm leading-relaxed text-card-foreground">
                "{t.quote}"
              </blockquote>
              <figcaption className="flex items-center gap-3 border-t border-border pt-4">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-primary to-accent text-sm font-semibold text-primary-foreground">
                  {t.initials}
                </div>
                <div>
                  <div className="text-sm font-semibold text-foreground">{t.name}</div>
                  <div className="text-xs text-muted-foreground">
                    {t.role} · {t.company}
                  </div>
                </div>
              </figcaption>
            </figure>
          ))}
        </div>
      </div>
    </section>
  );
}