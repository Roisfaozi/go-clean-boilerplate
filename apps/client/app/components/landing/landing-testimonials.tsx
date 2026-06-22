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
    <section
      id="testimonials"
      className="border-border bg-surface border-b py-20 md:py-24"
    >
      <div className="px-layout mx-auto max-w-7xl">
        <div className="mx-auto mb-14 max-w-2xl text-center">
          <span className="text-caption text-primary font-semibold tracking-wider uppercase">
            Testimonials
          </span>
          <h2 className="text-foreground mt-3 text-3xl font-bold tracking-tight md:text-4xl">
            Loved by teams building serious products
          </h2>
          <p className="text-body-lg text-muted-foreground mt-4">
            Bukti dari tim yang sudah membangun di atas NexusOS.
          </p>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {testimonials.map((t) => (
            <figure
              key={t.name}
              className="border-border bg-card relative flex flex-col rounded-xl border p-6"
            >
              <Quote className="text-primary/40 mb-4 h-6 w-6" />
              <blockquote className="text-card-foreground mb-6 flex-1 text-sm leading-relaxed">
                "{t.quote}"
              </blockquote>
              <figcaption className="border-border flex items-center gap-3 border-t pt-4">
                <div className="from-primary to-accent text-primary-foreground flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br text-sm font-semibold">
                  {t.initials}
                </div>
                <div>
                  <div className="text-foreground text-sm font-semibold">
                    {t.name}
                  </div>
                  <div className="text-muted-foreground text-xs">
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
