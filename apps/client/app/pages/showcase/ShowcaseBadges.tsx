import { PageHeader } from "@/components/layout/page-header";
import { NexusBadge } from "@casbin/ui";

export default function ShowcaseBadges() {
  return (
    <div className="max-w-5xl space-y-8">
      <PageHeader
        title="Badges"
        description="NexusBadge — semantic status indicators."
      />
      <div className="flex flex-wrap gap-3">
        <NexusBadge variant="success">Success</NexusBadge>
        <NexusBadge variant="warning">Warning</NexusBadge>
        <NexusBadge variant="danger">Danger</NexusBadge>
        <NexusBadge variant="info">Info</NexusBadge>
        <NexusBadge variant="neutral">Neutral</NexusBadge>
      </div>
    </div>
  );
}
