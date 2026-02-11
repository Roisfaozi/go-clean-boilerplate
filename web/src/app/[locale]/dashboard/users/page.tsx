"use client";

import { UserTable } from "~/components/dashboard/users/user-table";
import { UsersProvider, useUsers } from "./_components/users-context";
import { UsersHeader } from "./_components/users-header";
import { UsersToolbar } from "./_components/users-toolbar";
import { UsersPagination } from "./_components/users-pagination";
import { UsersModals } from "./_components/users-modals";
import { useMounted } from "~/hooks/use-mounted";

export default function UsersPage() {
  return (
    <UsersProvider>
      <UsersContent />
    </UsersProvider>
  );
}

function UsersContent() {
  const {
    users,
    isLoading,
    error,
    canUpdate,
    canDelete,
    handleEdit,
    handleDelete,
  } = useUsers();

  const isMounted = useMounted();

  return (
    <div className="space-y-4">
      <UsersHeader />
      <UsersToolbar />

      {/* Table */}
      <UserTable
        users={users}
        isLoading={isLoading}
        error={error}
        canUpdate={isMounted && canUpdate}
        canDelete={isMounted && canDelete}
        onEdit={handleEdit}
        onDelete={handleDelete}
      />

      <UsersPagination />
      <UsersModals />
    </div>
  );
}
