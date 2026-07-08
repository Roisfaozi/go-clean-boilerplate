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
		<section className="border-border bg-background border-b py-14">
			<div className="px-layout mx-auto max-w-7xl">
				<p className="text-caption text-muted-foreground text-center font-medium tracking-wider uppercase">
					Trusted by modern teams and enterprise operators
				</p>
				<div className="mt-8 grid grid-cols-2 items-center gap-8 sm:grid-cols-3 lg:grid-cols-6">
					{partners.map((name) => (
						<div
							key={name}
							className="text-body text-muted-foreground/70 hover:text-foreground flex items-center justify-center font-semibold tracking-tight transition-colors"
						>
							{name}
						</div>
					))}
				</div>
			</div>
		</section>
	);
}
