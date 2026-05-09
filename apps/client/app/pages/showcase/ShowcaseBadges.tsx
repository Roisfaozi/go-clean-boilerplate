import { PageHeader } from "@/components/layout/page-header";
import {  NexusBadge  } from "@casbin/ui";

export default function ShowcaseBadges() {
  return (
    <div className="space-y-8 max-w-5xl">
      <PageHeader title="Badges" description="NexusBadge — semantic status indicators." />
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
