"use client";

import { RolesProvider } from "./_components/roles-context";
import { RolesHeader } from "./_components/roles-header";
import { RolesGrid } from "./_components/roles-grid";
import { RolesModals } from "./_components/roles-modals";

export default function RolesPage() {
  return (
    <RolesProvider>
      <div className="space-y-6">
        <RolesHeader />
        <RolesGrid />
        <RolesModals />
      </div>
    </RolesProvider>
  );
}
