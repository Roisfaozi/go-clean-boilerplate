"use client";

import { UserTable } from "~/components/dashboard/users/user-table";
import { useMounted } from "~/hooks/use-mounted";
import { useUsers } from "./users-context";
import { UsersHeader } from "./users-header";
import { UsersModals } from "./users-modals";
import { UsersPagination } from "./users-pagination";
import { UsersToolbar } from "./users-toolbar";

export function UsersContent() {
  const {
    users,
    isLoading,
    error,
    searchTerm,
    canUpdate,
    canDelete,
    handleEdit,
    handleDelete,
    clearSearch,
    handleCreate,
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
        searchTerm={searchTerm}
        onClearSearch={clearSearch}
        onCreateUser={handleCreate}
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
