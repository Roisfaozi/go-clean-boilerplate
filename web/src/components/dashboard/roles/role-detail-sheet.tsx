"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "~/components/ui/sheet";
import { Role } from "~/lib/api/roles";
import { usersApi, User } from "~/lib/api/users";
import { accessApi } from "~/lib/api/access";
import { Icon } from "~/components/shared/icon";
import { Button } from "~/components/ui/button";
import { Badge } from "~/components/ui/badge";
import { ScrollArea } from "~/components/ui/scroll-area";
import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import { toast } from "sonner";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "~/components/ui/popover";
import { Skeleton } from "~/components/ui/skeleton";

interface RoleDetailSheetProps {
  role?: Role;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function RoleDetailSheet({
  role,
  open,
  onOpenChange,
}: RoleDetailSheetProps) {
  const [members, setMembers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isAdding, setIsAdding] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const fetchMembers = useCallback(async () => {
    if (!role) return;
    setIsLoading(true);
    try {
      // 1. Get user IDs from Casbin
      const resp = await accessApi.getUsersForRole(role.name);
      const userIds = resp.data || [];

      if (userIds.length === 0) {
        setMembers([]);
        return;
      }

      // 2. Fetch full user objects
      // Note: Backend might not have batch get users, so we might need to fetch one by one
      // or use search with filter if supported.
      // For now, let's try to fetch them individually or use a placeholder if many.
      const userPromises = userIds
        .slice(0, 50)
        .map((id) => usersApi.getById(id));
      const userResps = await Promise.all(userPromises);
      setMembers(userResps.map((r) => r.data).filter(Boolean) as User[]);
    } catch (error) {
      console.error("Failed to fetch role members", error);
      toast.error("Failed to load members");
    } finally {
      setIsLoading(false);
    }
  }, [role]);

  useEffect(() => {
    if (open && role) {
      fetchMembers();
    }
  }, [open, role, fetchMembers]);

  const handleSearch = async (query: string) => {
    setSearchQuery(query);
    if (query.length < 2) {
      setSearchResults([]);
      return;
    }

    setIsSearching(true);
    try {
      const resp = await usersApi.getAll(1, 10, query);
      if (resp.data) {
        // Filter out those who are already members
        const memberIds = new Set(members.map((m) => m.id));
        setSearchResults(resp.data.filter((u) => !memberIds.has(u.id)));
      }
    } catch (error) {
      console.error("Search failed", error);
    } finally {
      setIsSearching(false);
    }
  };

  const addMember = async (user: User) => {
    if (!role) return;
    try {
      await accessApi.assignRole(user.id, role.name);
      toast.success(`${user.username} added to ${role.name}`);
      setMembers((prev) => [...prev, user]);
      setSearchQuery("");
      setSearchResults([]);
    } catch (error) {
      toast.error("Failed to add member");
    }
  };

  const removeMember = async (userId: string, username: string) => {
    if (!role) return;
    try {
      await accessApi.revokeRole(userId, role.name);
      toast.success(`${username} removed from ${role.name}`);
      setMembers((prev) => prev.filter((m) => m.id !== userId));
    } catch (error) {
      toast.error("Failed to remove member");
    }
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex h-full flex-col sm:max-w-md">
        <SheetHeader>
          <div className="mb-2 flex items-center gap-2">
            <div className="bg-primary/10 text-primary rounded-md p-2">
              <Icon name="Shield" className="h-5 w-5" />
            </div>
            <SheetTitle className="text-xl">{role?.name}</SheetTitle>
          </div>
          <SheetDescription>
            {role?.description ||
              "Manage members and permissions for this role."}
          </SheetDescription>
        </SheetHeader>

        <div className="mt-6 flex min-h-0 flex-1 flex-col">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="flex items-center gap-2 text-sm font-semibold">
              <Icon name="Users" className="text-muted-foreground h-4 w-4" />
              Members ({members.length})
            </h3>

            <Popover>
              <PopoverTrigger asChild>
                <Button size="sm" variant="outline" className="h-8">
                  <Icon name="UserPlus" className="mr-2 h-4 w-4" />
                  Add Member
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-[300px] p-0" align="end">
                <Command shouldFilter={false}>
                  <CommandInput
                    placeholder="Search users..."
                    value={searchQuery}
                    onValueChange={handleSearch}
                  />
                  <CommandList>
                    {isSearching && (
                      <div className="p-4 text-center">
                        <Icon
                          name="Loader"
                          className="mx-auto mb-2 h-4 w-4 animate-spin"
                        />
                        <span className="text-muted-foreground text-xs">
                          Searching...
                        </span>
                      </div>
                    )}
                    {!isSearching &&
                      searchResults.length === 0 &&
                      searchQuery.length >= 2 && (
                        <CommandEmpty>No users found.</CommandEmpty>
                      )}
                    {!isSearching && searchQuery.length < 2 && (
                      <div className="text-muted-foreground p-4 text-center text-xs">
                        Type at least 2 characters to search...
                      </div>
                    )}
                    <CommandGroup>
                      {searchResults.map((user) => (
                        <CommandItem
                          key={user.id}
                          onSelect={() => addMember(user)}
                          className="flex cursor-pointer items-center gap-2"
                        >
                          <Avatar className="h-6 w-6">
                            <AvatarImage src={user.avatar_url} />
                            <AvatarFallback>
                              {user.username[0].toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <div className="flex flex-col">
                            <span className="text-sm font-medium">
                              {user.username}
                            </span>
                            <span className="text-muted-foreground text-[10px]">
                              {user.email}
                            </span>
                          </div>
                          <Icon
                            name="Plus"
                            className="text-muted-foreground ml-auto h-3 w-3"
                          />
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          </div>

          <ScrollArea className="-mx-2 flex-1 px-2">
            <div className="space-y-2">
              {isLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="flex items-center gap-3 p-2">
                    <Skeleton className="h-9 w-9 rounded-full" />
                    <div className="flex-1 space-y-1">
                      <Skeleton className="h-4 w-24" />
                      <Skeleton className="h-3 w-32" />
                    </div>
                  </div>
                ))
              ) : members.length === 0 ? (
                <div className="rounded-lg border-2 border-dashed py-12 text-center">
                  <Icon
                    name="Users"
                    className="text-muted-foreground/30 mx-auto mb-2 h-8 w-8"
                  />
                  <p className="text-muted-foreground text-sm">
                    No members assigned yet.
                  </p>
                </div>
              ) : (
                members.map((member) => (
                  <div
                    key={member.id}
                    className="hover:bg-muted/50 group flex items-center gap-3 rounded-md p-2 transition-colors"
                  >
                    <Avatar className="h-9 w-9 border">
                      <AvatarImage src={member.avatar_url} />
                      <AvatarFallback>
                        {member.username[0].toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0 flex-1">
                      <div className="mb-1 text-sm leading-none font-medium">
                        {member.username}
                      </div>
                      <div className="text-muted-foreground truncate text-xs">
                        {member.email}
                      </div>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="text-destructive hover:bg-destructive/10 h-8 w-8 opacity-0 group-hover:opacity-100"
                      onClick={() => removeMember(member.id, member.username)}
                    >
                      <Icon name="UserMinus" className="h-4 w-4" />
                    </Button>
                  </div>
                ))
              )}
            </div>
          </ScrollArea>
        </div>

        <div className="mt-auto border-t pt-6">
          <Button
            variant="outline"
            className="w-full"
            onClick={() => onOpenChange(false)}
          >
            Close
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}
