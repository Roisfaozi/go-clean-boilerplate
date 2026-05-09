export function LandingPartners() {
  const partners = [
    "Sentuh Digital",
    "Orion Ops",
    "CoreStack",
    "Helio Systems",
    "Atlas Group",
    "Karyon Labs",
  ];

  return (
    <section className="border-b border-border bg-background py-14">
      <div className="mx-auto max-w-7xl px-layout">
        <p className="text-center text-caption font-medium uppercase tracking-wider text-muted-foreground">
          Trusted by modern teams and enterprise operators
        </p>
        <div className="mt-8 grid grid-cols-2 items-center gap-8 sm:grid-cols-3 lg:grid-cols-6">
          {partners.map((name) => (
            <div
              key={name}
              className="flex items-center justify-center text-body font-semibold tracking-tight text-muted-foreground/70 transition-colors hover:text-foreground"
            >
              {name}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}