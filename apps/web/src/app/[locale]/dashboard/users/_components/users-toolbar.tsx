"use client";

import { useUsers } from "./users-context";
import { SearchInput } from "~/components/shared/search-input";
import { Button } from "~/components/ui/button";
import { Icon } from "~/components/shared/icon";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";

export function UsersToolbar() {
  const { searchTerm, handleSearch } = useUsers();

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 items-center space-x-2">
        <SearchInput
          defaultValue={searchTerm}
          onSearch={handleSearch}
          placeholder="Filter users..."
          className="h-8 w-[150px] lg:w-[250px]"
        />

        <Button variant="outline" size="sm" className="h-8 border-dashed">
          <Icon name="Plus" className="mr-2 h-4 w-4" />
          Status
        </Button>
        <Button variant="outline" size="sm" className="h-8 border-dashed">
          <Icon name="Plus" className="mr-2 h-4 w-4" />
          Role
        </Button>
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            className="ml-auto hidden h-8 lg:flex"
          >
            <Icon name="Settings" className="mr-2 h-4 w-4" />
            View
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-[150px]">
          <DropdownMenuLabel>Toggle columns</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuCheckboxItem checked>Avatar</DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem checked>Name</DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem checked>Email</DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem checked>Role</DropdownMenuCheckboxItem>
          <DropdownMenuCheckboxItem checked>Status</DropdownMenuCheckboxItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
