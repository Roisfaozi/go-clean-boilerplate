"use client";

import { useUsers } from "./users-context";
import { Button } from "~/components/ui/button";

export function UsersPagination() {
  const { page, limit, total, isLoading, handlePageChange } = useUsers();

  return (
    <div className="flex items-center justify-end space-x-2 py-4">
      <div className="text-muted-foreground flex-1 text-sm">
        Showing {(page - 1) * limit + 1} to {Math.min(page * limit, total)} of{" "}
        {total} results
      </div>
      <div className="space-x-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => handlePageChange(page - 1)}
          disabled={page === 1 || isLoading}
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => handlePageChange(page + 1)}
          disabled={page >= Math.ceil(total / limit) || isLoading}
        >
          Next
        </Button>
      </div>
    </div>
  );
}
