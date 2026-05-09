import { useState } from "react";
import { PageHeader } from "@/components/layout/page-header";
import { NexusButton } from "@casbin/ui";
import { Tooltip, TooltipContent, TooltipTrigger } from "@casbin/ui";
import { ConfirmationDialog } from "@casbin/ui";

export default function ShowcaseOverlays() {
  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <div className="max-w-5xl space-y-10">
      <PageHeader
        title="Overlays"
        description="Confirmation dialogs, tooltips, and other overlay components."
      />

      <section className="space-y-4">
        <h2 className="text-h2 text-foreground">Confirmation Dialog</h2>
        <NexusButton variant="danger" onClick={() => setConfirmOpen(true)}>
          Delete Item
        </NexusButton>
        <ConfirmationDialog
          open={confirmOpen}
          onOpenChange={setConfirmOpen}
          title="Delete this item?"
          description="This action cannot be undone."
          variant="danger"
          confirmLabel="Delete"
          onConfirm={() => setConfirmOpen(false)}
        />
      </section>

      <section className="space-y-4">
        <h2 className="text-h2 text-foreground">Tooltip</h2>
        <Tooltip>
          <TooltipTrigger asChild>
            <NexusButton variant="outline">Hover me</NexusButton>
          </TooltipTrigger>
          <TooltipContent>Tooltip content</TooltipContent>
        </Tooltip>
      </section>
    </div>
  );
}
