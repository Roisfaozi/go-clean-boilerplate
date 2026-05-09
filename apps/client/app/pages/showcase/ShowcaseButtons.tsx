import { PageHeader } from "@/components/layout/page-header";
import {  NexusButton  } from "@casbin/ui";
import {  NexusBadge  } from "@casbin/ui";
import { Info } from "lucide-react";

export default function ShowcaseButtons() {
  return (
    <div className="space-y-8 max-w-5xl">
      <PageHeader title="Buttons" description="NexusButton — all variants, sizes, and states." />

      <Section title="Variants">
        <div className="flex flex-wrap gap-3 items-center">
          <NexusButton variant="primary">Primary</NexusButton>
          <NexusButton variant="secondary">Secondary</NexusButton>
          <NexusButton variant="outline">Outline</NexusButton>
          <NexusButton variant="ghost">Ghost</NexusButton>
          <NexusButton variant="danger">Danger</NexusButton>
          <NexusButton variant="link">Link</NexusButton>
        </div>
      </Section>

      <Section title="Sizes">
        <div className="flex flex-wrap gap-3 items-center">
          <NexusButton size="sm">Small</NexusButton>
          <NexusButton size="default">Default</NexusButton>
          <NexusButton size="lg">Large</NexusButton>
          <NexusButton size="icon"><Info className="h-4 w-4" /></NexusButton>
        </div>
      </Section>

      <Section title="States">
        <div className="flex flex-wrap gap-3 items-center">
          <NexusButton disabled>Disabled</NexusButton>
          <NexusButton loading>Loading</NexusButton>
        </div>
      </Section>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return <section className="space-y-4"><h2 className="text-h2 text-foreground">{title}</h2>{children}</section>;
}
