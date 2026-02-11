"use client";

import { useOrganizationStore } from "~/stores/use-organization-store";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import { SettingsProvider, useSettings } from "./_components/settings-context";
import { GeneralSettingsCard } from "./_components/general-settings-card";
import { PreferencesSettingsCard } from "./_components/preferences-settings-card";
import { DangerZoneCard } from "./_components/danger-zone-card";

export default function OrganizationSettingsPage() {
  const { currentOrganization } = useOrganizationStore();

  if (!currentOrganization) {
    return (
      <div className="flex h-[400px] items-center justify-center rounded-lg border-2 border-dashed">
        <p className="text-muted-foreground">No organization selected.</p>
      </div>
    );
  }

  return (
    <SettingsProvider>
      <OrganizationSettingsContent />
    </SettingsProvider>
  );
}

function OrganizationSettingsContent() {
  const { handleUpdate, isLoading, hasChanges } = useSettings();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">
            Organization Settings
          </h2>
          <p className="text-muted-foreground">
            Update your organization profile and general settings.
          </p>
        </div>
        <Button onClick={handleUpdate} disabled={isLoading || !hasChanges}>
          {isLoading && (
            <Icon name="Loader" className="mr-2 h-4 w-4 animate-spin" />
          )}
          Save All Changes
        </Button>
      </div>

      <div className="grid max-w-3xl gap-6">
        <GeneralSettingsCard />
        <PreferencesSettingsCard />
        <DangerZoneCard />
      </div>
    </div>
  );
}
