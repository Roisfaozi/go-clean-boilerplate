"use client";

import { AuditProvider } from "./_components/audit-context";
import { AuditToolbar } from "./_components/audit-toolbar";
import { AuditTable } from "./_components/audit-table";
import { AuditPagination } from "./_components/audit-pagination";

export default function AuditPage() {
  return (
    <AuditProvider>
      <div className="space-y-4">
        <AuditToolbar />
        <AuditTable />
        <AuditPagination />
      </div>
    </AuditProvider>
  );
}
