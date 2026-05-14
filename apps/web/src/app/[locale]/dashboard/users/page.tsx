import { usersApi } from "~/lib/api/users";
import { UsersContent } from "./_components/users-content";
import { UsersProvider } from "./_components/users-context";

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
