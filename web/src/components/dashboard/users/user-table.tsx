"use client";

import { format } from "date-fns";
import Image from "next/image";
import { Icon } from "~/components/shared/icon";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "~/components/ui/table";
import { User } from "~/lib/api/users";
import { memo } from "react";
import { EmptyState } from "~/components/shared/empty-state";

interface UserTableProps {
  users: User[];
  isLoading: boolean;
  error: any;
  canUpdate: boolean;
  canDelete: boolean;
  onEdit: (user: User) => void;
  onDelete: (user: User) => void;
  searchTerm?: string;
  onClearSearch?: () => void;
  onCreateUser?: () => void;
}

export function UserTable({
  users,
  isLoading,
  error,
  canUpdate,
  canDelete,
  onEdit,
  onDelete,
  searchTerm,
  onClearSearch,
  onCreateUser,
}: UserTableProps) {
  if (!isLoading && !error && users.length === 0) {
    return (
      <div className="bg-muted/5 rounded-md border border-dashed">
        {searchTerm ? (
          <EmptyState
            case="search"
            searchTerm={searchTerm}
            action={
              onClearSearch
                ? { label: "Clear search", onClick: onClearSearch }
                : undefined
            }
          />
        ) : (
          <EmptyState
            case="users"
            action={
              onCreateUser
                ? {
                    label: "Add Your First User",
                    onClick: onCreateUser,
                    icon: "UserPlus",
                  }
                : undefined
            }
          />
        )}
      </div>
    );
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[50px]"></TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Username</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="text-right">Joined</TableHead>
            {(canUpdate || canDelete) && (
              <TableHead className="w-[50px]"></TableHead>
            )}
          </TableRow>
        </TableHeader>
        <TableBody>
          {isLoading ? (
            <TableRow>
              <TableCell colSpan={7} className="h-24 text-center">
                <div className="flex items-center justify-center gap-2">
                  <Icon name="Loader" className="h-4 w-4 animate-spin" />
                  Loading...
                </div>
              </TableCell>
            </TableRow>
          ) : error ? (
            <TableRow>
              <TableCell colSpan={7} className="h-24 text-center">
                <EmptyState
                  case="error"
                  action={{
                    label: "Retry",
                    onClick: () => window.location.reload(),
                  }}
                />
              </TableCell>
            </TableRow>
          ) : (
            users.map((user) => (
              <MemoizedUserTableRow
                key={user.id}
                user={user}
                canUpdate={canUpdate}
                canDelete={canDelete}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}

const MemoizedUserTableRow = memo(function UserTableRow({
  user,
  canUpdate,
  canDelete,
  onEdit,
  onDelete,
}: {
  user: User;
  canUpdate: boolean;
  canDelete: boolean;
  onEdit: (user: User) => void;
  onDelete: (user: User) => void;
}) {
  return (
    <TableRow>
      <TableCell>
        <div className="bg-muted flex h-8 w-8 items-center justify-center overflow-hidden rounded-full text-xs font-medium">
          {user.avatar_url ? (
            <Image
              src={user.avatar_url}
              alt={user.name}
              width={32}
              height={32}
              className="h-full w-full object-cover"
            />
          ) : (
            user.name.charAt(0).toUpperCase()
          )}
        </div>
      </TableCell>
      <TableCell className="font-medium">{user.name}</TableCell>
      <TableCell>{user.email}</TableCell>
      <TableCell>{user.username}</TableCell>
      <TableCell>
        <Badge
          variant={user.status === "active" ? "default" : "secondary"}
          className={
            user.status === "active"
              ? "bg-emerald-500 hover:bg-emerald-600"
              : ""
          }
        >
          {user.status || "Unknown"}
        </Badge>
      </TableCell>
      <TableCell className="text-muted-foreground text-right">
        {format(new Date(user.created_at * 1000), "MMM dd, yyyy")}
      </TableCell>
      {(canUpdate || canDelete) && (
        <TableCell>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <Icon name="Ellipsis" className="h-4 w-4" />
                <span className="sr-only">Open menu</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {canUpdate && (
                <DropdownMenuItem onClick={() => onEdit(user)}>
                  <Icon name="Pencil" className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
              )}
              {canDelete && (
                <DropdownMenuItem
                  onClick={() => onDelete(user)}
                  className="text-destructive"
                >
                  <Icon name="Trash" className="mr-2 h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </TableCell>
      )}
    </TableRow>
  );
});
