import { UserTable } from "~/components/dashboard/users/user-table";
import { UsersProvider, useUsers } from "./_components/users-context";
import { UsersHeader } from "./_components/users-header";
import { UsersToolbar } from "./_components/users-toolbar";
import { UsersPagination } from "./_components/users-pagination";
import { UsersModals } from "./_components/users-modals";
import { useMounted } from "~/hooks/use-mounted";
import { usersApi } from "~/lib/api/users";

export default async function UsersPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; limit?: string; search?: string }>;
}) {
  const resolvedParams = await searchParams;
  const page = Number(resolvedParams.page) || 1;
  const limit = Number(resolvedParams.limit) || 10;
  const search = resolvedParams.search || "";

  // 1. Fetch data on Server (Critical Path)
  const initialData = await usersApi.getAll(page, limit, search);

  return (
    <UsersProvider initialData={initialData}>
      <UsersContent />
    </UsersProvider>
  );
}

function UsersContent() {
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
